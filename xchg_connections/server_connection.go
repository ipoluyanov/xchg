package xchg_connections

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"sync"
	"time"

	"github.com/ipoluianov/gomisc/crypt_tools"
	"github.com/ipoluianov/gomisc/logger"
	"github.com/ipoluianov/gomisc/nonce_generator"
	"github.com/ipoluianov/gomisc/snake_counter"
	"github.com/ipoluianov/xchg/xchg_network"
)

type Session struct {
	id           uint64
	aesKey       []byte
	lastAccessDT time.Time
	snakeCounter *snake_counter.SnakeCounter
}

type ServerConnection struct {
	mtxServerConnection sync.Mutex
	edgeConnections     map[string]*PeerConnection
	privateKey32        string
	privateKey          *rsa.PrivateKey

	sessionsById          map[uint64]*Session
	nextSessionId         uint64
	lastPurgeSessionsTime time.Time
	network               *xchg_network.Network

	nonceGenerator *nonce_generator.NonceGenerator

	processor ServerProcessor
}

type ServerProcessor interface {
	ServerProcessorAuth(authData []byte) (err error)
	ServerProcessorCall(function string, parameter []byte) (response []byte, err error)
}

func NewServerConnection(privateKey32 string, network *xchg_network.Network) *ServerConnection {
	//logger.Println("[i]", "NewServerConnection", "begin", network.String())
	var err error
	var c ServerConnection
	c.nextSessionId = 1
	c.nonceGenerator = nonce_generator.NewNonceGenerator(1024 * 1024)
	c.privateKey32 = privateKey32
	privateKeyBS, err := base32.StdEncoding.DecodeString(c.privateKey32)
	if err != nil {
		logger.Println("[ERROR]", "NewServerConnection", "base32.StdEncoding.DecodeString error:", err)
	}
	c.privateKey, err = crypt_tools.RSAPrivateKeyFromDer(privateKeyBS)
	if err != nil {
		logger.Println("[ERROR]", "NewServerConnection", "crypt_tools.RSAPrivateKeyFromDer error:", err)
	}
	c.edgeConnections = make(map[string]*PeerConnection)
	c.sessionsById = make(map[uint64]*Session)
	c.network = network
	publicKeyBS := crypt_tools.RSAPublicKeyToDer(&c.privateKey.PublicKey)
	//logger.Println("[i]", "NewServerConnection", "creating nodes", publicKeyBS)
	addresses := c.network.GetAddressesByPublicKey(publicKeyBS)
	connections := make([]*PeerConnection, 0)
	//logger.Println("[i]", "NewServerConnection", "creating nodes", addresses)
	for _, addr := range addresses {
		conn := c.connection(addr)
		connections = append(connections, conn)
		//logger.Println("[i]", "NewServerConnection", "added connection:", addr)
	}
	//logger.Println("[i]", "NewServerConnection", "end")
	return &c
}

func (c *ServerConnection) SetProcessor(processor ServerProcessor) {
	c.processor = processor
}

func (c *ServerConnection) SetNetwork(network *xchg_network.Network) {
	c.network = network
}

func (c *ServerConnection) Start() {
}

func (c *ServerConnection) connection(addr string) *PeerConnection {
	c.mtxServerConnection.Lock()
	defer c.mtxServerConnection.Unlock()
	if conn, ok := c.edgeConnections[addr]; ok {
		return conn
	}
	conn := NewPeerConnection(addr, c.privateKey)
	conn.SetProcessor(c)
	c.edgeConnections[addr] = conn
	conn.Start()
	return conn
}

func (c *ServerConnection) purgeSessions() {
	now := time.Now()
	c.mtxServerConnection.Lock()
	if now.Sub(c.lastPurgeSessionsTime).Seconds() > 60 {
		for sessionId, session := range c.sessionsById {
			if now.Sub(session.lastAccessDT).Seconds() > 60 {
				delete(c.sessionsById, sessionId)
			}
		}
		c.lastPurgeSessionsTime = time.Now()
	}
	c.mtxServerConnection.Unlock()
}

func (c *ServerConnection) OnEdgeConnected(edgeConnection *PeerConnection) {
}

func (c *ServerConnection) OnEdgeDissonnected(edgeConnection *PeerConnection) {
}

func (c *ServerConnection) OnEdgeReceivedCall(edgeConnection *PeerConnection, sessionId uint64, data []byte) (response []byte) {

	var err error
	// Find the session
	var session *Session
	if sessionId != 0 {
		c.mtxServerConnection.Lock()
		var ok bool
		session, ok = c.sessionsById[sessionId]
		if !ok {
			session = nil
		}
		c.mtxServerConnection.Unlock()
	}

	if sessionId != 0 {
		if session == nil {
			response = c.prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_SESSION))
			return
		}
		data, err = crypt_tools.DecryptAESGCM(data, session.aesKey)
		if err != nil {
			response = c.prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_DECR + ":" + err.Error()))
			return
		}
		if len(data) < 9 {
			response = c.prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_LEN9))
			return
		}
		callNonce := binary.LittleEndian.Uint64(data)
		err = session.snakeCounter.TestAndDeclare(int(callNonce))
		if err != nil {
			response = c.prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_NONCE))
			return
		}
		data = data[8:]
	} else {
		if len(data) < 1 {
			response = c.prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_LEN1))
			return
		}
	}

	functionLen := int(data[0])
	if len(data) < 1+functionLen {
		response = c.prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_LEN_FN))
		return
	}
	function := string(data[1 : 1+functionLen])
	functionParameter := data[1+functionLen:]

	var processor ServerProcessor
	c.mtxServerConnection.Lock()
	processor = c.processor
	c.mtxServerConnection.Unlock()

	if processor == nil {
		response = c.prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_NOT_IMPL))
		return
	}

	var resp []byte

	switch function {
	case "/xchg-get-nonce":
		resp = c.nonceGenerator.Get()
	case "/xchg-auth":
		resp, err = c.processAuth(functionParameter)
	default:
		resp, err = c.processor.ServerProcessorCall(function, functionParameter)
	}

	if err != nil {
		response = c.prepareResponseError(err)
	} else {
		response = c.prepareResponse(resp)
	}
	return
}

func (c *ServerConnection) processAuth(functionParameter []byte) (response []byte, err error) {
	if len(functionParameter) < 4 {
		err = errors.New(ERR_XCHG_SRV_CONN_AUTH_DATA_LEN4)
		return
	}

	remotePublicKeyBSLen := binary.LittleEndian.Uint32(functionParameter)
	if len(functionParameter) < 4+int(remotePublicKeyBSLen) {
		err = errors.New(ERR_XCHG_SRV_CONN_AUTH_DATA_LEN_PK)
		return
	}

	remotePublicKeyBS := functionParameter[4 : 4+remotePublicKeyBSLen]
	var remotePublicKey *rsa.PublicKey
	remotePublicKey, err = crypt_tools.RSAPublicKeyFromDer(remotePublicKeyBS)
	if err != nil {
		return
	}

	authFrameSecret := functionParameter[4+remotePublicKeyBSLen:]

	var parameter []byte
	parameter, err = rsa.DecryptPKCS1v15(rand.Reader, c.privateKey, authFrameSecret)
	if err != nil {
		return
	}

	if len(parameter) < 16 {
		err = errors.New(ERR_XCHG_SRV_CONN_AUTH_DATA_LEN_NONCE)
		return
	}

	nonce := parameter[0:16]
	if !c.nonceGenerator.Check(nonce) {
		err = errors.New(ERR_XCHG_SRV_CONN_AUTH_WRONG_NONCE)
		return
	}

	authData := parameter[16:]
	err = c.processor.ServerProcessorAuth(authData)
	if err != nil {
		return
	}

	c.mtxServerConnection.Lock()

	sessionId := c.nextSessionId
	c.nextSessionId++
	session := &Session{}
	session.id = sessionId
	session.lastAccessDT = time.Now()
	session.aesKey = make([]byte, 32)
	session.snakeCounter = snake_counter.NewSnakeCounter(100, 0)
	rand.Read(session.aesKey)
	c.sessionsById[sessionId] = session
	response = make([]byte, 8+32)
	binary.LittleEndian.PutUint64(response, sessionId)
	copy(response[8:], session.aesKey)
	response, err = rsa.EncryptPKCS1v15(rand.Reader, remotePublicKey, response)

	c.mtxServerConnection.Unlock()

	return
}

func (c *ServerConnection) prepareResponse(data []byte) []byte {
	respFrame := make([]byte, 1+len(data))
	respFrame[0] = 0
	copy(respFrame[1:], data)
	return respFrame
}

func (c *ServerConnection) prepareResponseError(err error) []byte {
	errBS := make([]byte, 0)
	if err != nil {
		errBS = []byte(err.Error())
	}
	respFrame := make([]byte, 1+len(errBS))
	respFrame[0] = 1
	copy(respFrame[1:], errBS)
	return respFrame
}

const (
	ERR_XCHG_SRV_CONN_WRONG_SESSION       = "{ERR_XCHG_SRV_CONN_WRONG_SESSION}"
	ERR_XCHG_SRV_CONN_DECR                = "{ERR_XCHG_SRV_CONN_DECR}"
	ERR_XCHG_SRV_CONN_WRONG_LEN9          = "{ERR_XCHG_SRV_CONN_WRONG_LEN9}"
	ERR_XCHG_SRV_CONN_WRONG_NONCE         = "{ERR_XCHG_SRV_CONN_WRONG_NONCE}"
	ERR_XCHG_SRV_CONN_WRONG_LEN1          = "{ERR_XCHG_SRV_CONN_WRONG_LEN1}"
	ERR_XCHG_SRV_CONN_WRONG_LEN_FN        = "{ERR_XCHG_SRV_CONN_WRONG_LEN_FN}"
	ERR_XCHG_SRV_CONN_AUTH_DATA_LEN4      = "{ERR_XCHG_SRV_CONN_AUTH_DATA_LEN4}"
	ERR_XCHG_SRV_CONN_NOT_IMPL            = "{ERR_XCHG_SRV_CONN_NOT_IMPL}"
	ERR_XCHG_SRV_CONN_AUTH_DATA_LEN_NONCE = "{ERR_XCHG_SRV_CONN_AUTH_DATA_LEN_NONCE}"
	ERR_XCHG_SRV_CONN_AUTH_DATA_LEN_PK    = "{ERR_XCHG_SRV_CONN_AUTH_DATA_LEN_PK}"
	ERR_XCHG_SRV_CONN_AUTH_WRONG_NONCE    = "{ERR_XCHG_SRV_CONN_AUTH_WRONG_NONCE}"
)
