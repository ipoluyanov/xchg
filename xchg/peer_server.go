package xchg

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"errors"
	"log"
	"time"
)

func (c *Peer) onEdgeReceivedCall(sessionId uint64, data []byte) (response []byte, dontSendResponse bool) {

	var err error
	// Find the session
	var session *Session
	encryped := false
	if sessionId != 0 {
		c.mtx.Lock()
		var ok bool
		session, ok = c.sessionsById[sessionId]
		if !ok {
			session = nil
		}
		c.mtx.Unlock()
	}

	if sessionId != 0 {
		if session == nil {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_SESSION))
			return
		}
		data, err = DecryptAESGCM(data, session.aesKey)
		if err != nil {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_DECR + ":" + err.Error()))
			return
		}
		data, err = UnpackBytes(data)
		if err != nil {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_UNPACK + ":" + err.Error()))
			return
		}
		if len(data) < 9 {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_LEN9))
			return
		}
		encryped = true
		callNonce := binary.LittleEndian.Uint64(data)
		err = session.snakeCounter.TestAndDeclare(int(callNonce))
		if err != nil {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_NONCE))
			dontSendResponse = true
			return
		}
		data = data[8:]
		session.lastAccessDT = time.Now()
	} else {
		if len(data) < 1 {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_LEN1))
			return
		}
	}

	functionLen := int(data[0])
	if len(data) < 1+functionLen {
		response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_LEN_FN))
		return
	}
	function := string(data[1 : 1+functionLen])
	functionParameter := data[1+functionLen:]

	var processor ServerProcessor
	c.mtx.Lock()
	processor = c.processor
	c.mtx.Unlock()

	if processor == nil {
		response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_NOT_IMPL))
		return
	}

	var resp []byte

	switch function {
	case "/xchg-get-nonce":
		nonce := c.authNonces.Next()
		resp = nonce[:]
	case "/xchg-auth":
		resp, err = c.processAuth(functionParameter)
	default:
		resp, err = c.processor.ServerProcessorCall(function, functionParameter)
	}

	if err != nil {
		response = prepareResponseError(err)
	} else {
		response = prepareResponse(resp)
	}

	if encryped {
		response = PackBytes(response)
		response, err = EncryptAESGCM(response, session.aesKey)
		if err != nil {
			return
		}
	}

	return
}

func (c *Peer) processAuth(functionParameter []byte) (response []byte, err error) {
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
	remotePublicKey, err = RSAPublicKeyFromDer(remotePublicKeyBS)
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
	if !c.authNonces.Check(nonce) {
		err = errors.New(ERR_XCHG_SRV_CONN_AUTH_WRONG_NONCE)
		return
	}

	authData := parameter[16:]
	err = c.processor.ServerProcessorAuth(authData)
	if err != nil {
		return
	}

	c.mtx.Lock()

	sessionId := c.nextSessionId
	c.nextSessionId++
	session := &Session{}
	session.id = sessionId
	session.lastAccessDT = time.Now()
	session.aesKey = make([]byte, 32)
	session.snakeCounter = NewSnakeCounter(100, 0)
	rand.Read(session.aesKey)
	c.sessionsById[sessionId] = session
	response = make([]byte, 8+32)
	binary.LittleEndian.PutUint64(response, sessionId)
	copy(response[8:], session.aesKey)
	response, err = rsa.EncryptPKCS1v15(rand.Reader, remotePublicKey, response)

	c.mtx.Unlock()

	return
}

func (c *Peer) purgeSessions() {
	now := time.Now()
	c.mtx.Lock()
	if now.Sub(c.lastPurgeSessionsTime).Seconds() > 60 {
		for sessionId, session := range c.sessionsById {
			if now.Sub(session.lastAccessDT).Seconds() > 60 {
				delete(c.sessionsById, sessionId)
				log.Println("Session removed", sessionId)
			}
		}
		c.lastPurgeSessionsTime = time.Now()
	}
	c.mtx.Unlock()
}

func prepareResponseError(err error) []byte {
	//fmt.Println("CALL ERROR:", err)

	errBS := make([]byte, 0)
	if err != nil {
		errBS = []byte(err.Error())
	}
	respFrame := make([]byte, 1+len(errBS))
	respFrame[0] = 1
	copy(respFrame[1:], errBS)
	return respFrame
}

func prepareResponse(data []byte) []byte {
	respFrame := make([]byte, 1+len(data))
	respFrame[0] = 0
	copy(respFrame[1:], data)
	return respFrame
}
