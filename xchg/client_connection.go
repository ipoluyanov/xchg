package xchg

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"errors"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ipoluianov/gomisc/crypt_tools"
)

type ClientConnection struct {
	edgeConnections map[string]*EdgeConnection
	address         string
	remotePublicKey *rsa.PublicKey
	localPrivateKey *rsa.PrivateKey

	aesKey              []byte
	sessionId           uint64
	sessionNonceCounter uint64

	secretBytes []byte

	authData string
}

func NewClientConnection(address string, localPrivateKey58 string, authData string) *ClientConnection {
	var c ClientConnection
	c.address = address
	c.authData = authData
	c.remotePublicKey, _ = crypt_tools.RSAPublicKeyFromDer(base58.Decode(address))
	c.edgeConnections = make(map[string]*EdgeConnection)
	c.localPrivateKey, _ = crypt_tools.RSAPrivateKeyFromDer(base58.Decode(localPrivateKey58))
	eConn := NewEdgeConnection("localhost:8484", c.localPrivateKey)
	c.edgeConnections["localhost:8484"] = eConn
	c.sessionNonceCounter = 1

	eConn.Start()
	return &c
}

func (c *ClientConnection) getEdgeConnection() *EdgeConnection {
	return c.edgeConnections["localhost:8484"]
}

func (c *ClientConnection) Call(function string, data []byte) (result []byte, err error) {
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
	var nonce []byte
	nonce, err = c.regularCall("/xchg-get-nonce", nil, nil)
	if err != nil {
		return
	}
	if len(nonce) != 16 {
		err = errors.New("wrong nonce len")
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
		err = errors.New("wrong auth response")
		return
	}
	c.sessionId = binary.LittleEndian.Uint64(result)
	c.aesKey = make([]byte, 32)
	copy(c.aesKey, result[8:])
	return
}

func (c *ClientConnection) regularCall(function string, data []byte, aesKey []byte) (result []byte, err error) {
	if len(function) > 255 {
		err = errors.New("wrong function len")
		return
	}
	ec := c.getEdgeConnection()
	if ec == nil {
		err = errors.New("no connection")
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
			err = errors.New("client_connection encrypt error: " + err.Error())
			return
		}
	} else {
		frame = make([]byte, 1+len(function)+len(data))
		frame[0] = byte(len(function))
		copy(frame[1:], function)
		copy(frame[1+len(function):], data)
	}

	result, err = ec.Call(c.address, c.sessionId, frame)
	if err != nil {
		return
	}

	if len(result) < 1 {
		err = errors.New("wrong response (<1)")
		return
	}

	if result[0] == 0 {
		result = result[1:]
		err = nil
		return
	}

	if result[0] == 1 {
		err = errors.New(string(result[1:]))
		result = nil
		return
	}

	err = errors.New("wrong response (status byte)")
	return
}
