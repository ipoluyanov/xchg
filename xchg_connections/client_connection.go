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
	"github.com/ipoluianov/xchg/xchg"
	"github.com/ipoluianov/xchg/xchg_network"
)

type ClientConnection struct {
	mtxClientConnection sync.Mutex
	address             string
	remotePublicKey     *rsa.PublicKey
	localPrivateKey     *rsa.PrivateKey

	aesKey              []byte
	sessionId           uint64
	sessionNonceCounter uint64

	findingConnection bool
	currentConnection *PeerConnection
	currentSID        uint64

	secretBytes []byte

	network *xchg_network.Network

	authData string

	lastestNodeAddress string
	authProcessing     bool

	onEvent FuncOnEvent
}

type FuncOnEvent func(text string)

type ClientConnectionState struct {
	SessionId           uint64
	FindingConnection   bool
	CurrentSID          uint64
	AuthProcessing      bool
	LatestNodeAddress   string
	PeerConnectionState PeerConnectionState
}

func NewClientConnection(network *xchg_network.Network, address string, localPrivateKey32 string, authData string, onEvent FuncOnEvent) *ClientConnection {
	var c ClientConnection
	c.address = address
	c.authData = authData
	c.network = network
	c.remotePublicKey = nil
	localPrivateKeyBS, _ := base32.StdEncoding.DecodeString(localPrivateKey32)
	c.localPrivateKey, _ = crypt_tools.RSAPrivateKeyFromDer(localPrivateKeyBS)
	c.sessionNonceCounter = 1
	c.onEvent = onEvent
	return &c
}

func (c *ClientConnection) Dispose() {
	c.mtxClientConnection.Lock()
	defer c.mtxClientConnection.Unlock()
	c.onEvent = nil
	if c.currentConnection != nil {
		c.currentConnection.Dispose()
		c.currentConnection = nil
	}
}

func (c *ClientConnection) Reset() {
	c.mtxClientConnection.Lock()
	c.reset()
	c.mtxClientConnection.Unlock()
}

func (c *ClientConnection) reset() {
	c.currentSID = 0
	c.sessionId = 0
	if c.currentConnection != nil {
		c.currentConnection.Stop()
		c.currentConnection.Dispose()
		c.currentConnection = nil
	}
}

func (c *ClientConnection) CallOnEvent(text string) {
	c.mtxClientConnection.Lock()
	onEvent := c.onEvent
	c.mtxClientConnection.Unlock()
	if onEvent != nil {
		onEvent(text)
	}
}

func (c *ClientConnection) Call(function string, data []byte) (result []byte, err error) {
	c.CallOnEvent("call " + function)
	if c.sessionId == 0 {
		err = c.auth()
		if err != nil {
			return
		}
	}
	result, err = c.regularCall(function, data, c.aesKey)
	return
}

func (c *ClientConnection) auth() (err error) {
	c.authProcessing = true
	defer func() {
		c.authProcessing = false
	}()

	var nonce []byte
	nonce, err = c.regularCall("/xchg-get-nonce", nil, nil)
	if err != nil {
		return
	}
	if len(nonce) != 16 {
		err = errors.New(xchg.ERR_XCHG_CL_CONN_WRONG_NONCE_LEN)
		return
	}

	localPublicKeyBS := crypt_tools.RSAPublicKeyToDer(&c.localPrivateKey.PublicKey)

	authFrameSecret := make([]byte, 16+len(c.authData))
	copy(authFrameSecret, nonce)
	copy(authFrameSecret[16:], []byte(c.authData))

	var encryptedAuthFrame []byte
	encryptedAuthFrame, err = rsa.EncryptPKCS1v15(rand.Reader, c.remotePublicKey, []byte(authFrameSecret))
	if err != nil {
		return
	}

	authFrame := make([]byte, 4+len(localPublicKeyBS)+len(encryptedAuthFrame))
	binary.LittleEndian.PutUint32(authFrame, uint32(len(localPublicKeyBS)))
	copy(authFrame[4:], localPublicKeyBS)
	copy(authFrame[4+len(localPublicKeyBS):], encryptedAuthFrame)

	var result []byte
	result, err = c.regularCall("/xchg-auth", authFrame, nil)
	if err != nil {
		return
	}

	result, err = rsa.DecryptPKCS1v15(rand.Reader, c.localPrivateKey, result)
	if err != nil {
		return
	}

	if len(result) != 8+32 {
		err = errors.New(xchg.ERR_XCHG_CL_CONN_WRONG_AUTH_RESPONSE_LEN)
		return
	}
	c.sessionId = binary.LittleEndian.Uint64(result)
	c.aesKey = make([]byte, 32)
	copy(c.aesKey, result[8:])
	return
}

func (c *ClientConnection) regularCall(function string, data []byte, aesKey []byte) (result []byte, err error) {
	if len(function) > 255 {
		err = errors.New(xchg.ERR_XCHG_CL_CONN_WRONG_FUNCTION_LEN)
		return
	}

	if c.findingConnection {
		err = errors.New(xchg.ERR_XCHG_CL_CONN_SEARCHING_NODE)
		return
	}

	c.mtxClientConnection.Lock()
	c.findingConnection = true
	if c.currentSID == 0 {
		c.reset()
		//logger.Println("[i]", "ClientConnection::regularCall", "searching node ...")
		addresses := c.network.GetAddressesByPublicKey(crypt_tools.RSAPublicKeyToDer(c.remotePublicKey))
		for _, address := range addresses {
			c.lastestNodeAddress = address
			//logger.Println("[i]", "ClientConnection::regularCall", "trying node:", address)

			conn := NewPeerConnection(address, c.localPrivateKey, nil)
			conn.Start()
			if !conn.WaitForConnection(500 * time.Millisecond) {
				conn.Stop()
				continue
			}

			c.currentSID, c.remotePublicKey, err = conn.ResolveAddress(c.address)
			if c.currentSID != 0 {
				// Check public key
				if xchg.AddressForPublicKey(c.remotePublicKey) == c.address {
					c.currentConnection = conn
					//logger.Println("[i]", "ClientConnection::regularCall", "node found:", address)
					break
				}
			}

			conn.Stop()
			conn.Dispose()
		}
	}

	connection := c.currentConnection
	currentSID := c.currentSID
	sessionId := c.sessionId
	c.findingConnection = false
	c.mtxClientConnection.Unlock()

	if connection == nil || currentSID == 0 {
		err = errors.New(xchg.ERR_XCHG_CL_CONN_NO_ROUTE_TO_PEER)
		return
	}

	var frame []byte
	if len(aesKey) == 32 {
		frame = make([]byte, 8+1+len(function)+len(data))
		binary.LittleEndian.PutUint64(frame, c.sessionNonceCounter)
		c.sessionNonceCounter++
		frame[8] = byte(len(function))
		copy(frame[9:], function)
		copy(frame[9+len(function):], data)
		frame, err = crypt_tools.EncryptAESGCM(frame, aesKey)
		if err != nil {
			c.Reset()
			err = errors.New(xchg.ERR_XCHG_CL_CONN_CALL_ENC + ":" + err.Error())
			return
		}
	} else {
		frame = make([]byte, 1+len(function)+len(data))
		frame[0] = byte(len(function))
		copy(frame[1:], function)
		copy(frame[1+len(function):], data)
	}

	result, err = connection.Call(currentSID, sessionId, frame)

	if xchg.NeedToChangeNode(err) {
		c.Reset()
		return
	}

	if err != nil {
		err = errors.New(xchg.ERR_XCHG_CL_CONN_CALL_ERR + ":" + err.Error())
		return
	}

	if len(result) < 1 {
		err = errors.New(xchg.ERR_XCHG_CL_CONN_WRONG_CALL_RESPONSE)
		c.Reset()
		return
	}

	if result[0] == 0 {
		// Success response
		result = result[1:]
		err = nil
		return
	}

	if result[0] == 1 {
		// Error response
		err = errors.New(xchg.ERR_XCHG_CL_CONN_FROM_PEER + ":" + string(result[1:]))
		if xchg.NeedToMakeSession(err) {
			// Any server error - make new session
			c.sessionId = 0
		}
		result = nil
		return
	}

	err = errors.New(xchg.ERR_XCHG_CL_CONN_WRONG_CALL_RESPONSE_BYTE)
	c.Reset()
	return
}

func (c *ClientConnection) State() ClientConnectionState {
	c.mtxClientConnection.Lock()
	defer c.mtxClientConnection.Unlock()

	var state ClientConnectionState
	state.CurrentSID = c.currentSID
	state.AuthProcessing = c.authProcessing
	state.FindingConnection = c.findingConnection
	state.SessionId = c.sessionId
	state.LatestNodeAddress = c.lastestNodeAddress

	if c.currentConnection != nil {
		state.PeerConnectionState = c.currentConnection.State()
	}

	return state
}
