package xchg

import (
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	NetworkContainerDefault             = "UEsDBBQACAAIAFFxdFUAAAAAAAAAAAAAAAAMAAkAbmV0d29yay5qc29uVVQFAAFLJ3pj1NBPauswEAbwvU4xzDooif/bd3jvAsUUxZ7YorZkZJWahNy9KHULDYV4OxstRh8fv5mrADRqJKwA/ylt/pPHnQD0eqTZq3HCCo5ZVpRJWcRp+NFGe62G18lq42es4EUAYO/9VO33S9N3izTk94b8h3Vv8qIneVIzZUno/RWUnbqQk81g39tneTUM1JBs7LgxucGwJq3r/koKqMO6TpmOvte8CgDAydFZL+Fih3sfYG/nn1OsKQBUbetoDnNMExnnMo/lMTtUxaH4goTQ4/Hv49vuaVEUlVuKwluLtfGRf+TNj3jzY978hDc/5c3PePNz3vyCN7/kzVe8+Sfe/IY3v+XNJ978Mx++gFrcPgMAAP//UEsHCHij418RAQAAHw0AAFBLAwQUAAgACABRcXRVAAAAAAAAAAAAAAAAEAAJAHNpZ25hdHVyZS5iYXNlNjRVVAUAAUsnemMEwMmCsiAAAOAH6iA02m+HOZSJC6hZuM0tc8U1VEyf/v84gfjarMfDx52PfVcd9wkLNX3Qyht2ot4ggemwyAR2AorvKR3W3uZenZbadPuHEreLrA1z8Pfcwc+kEV7nvkwOtS6NKtZP50qjgSqGvLGYabpvSwE4fF36gbIF3bORlsdzsEogOH9963ppB4085T6ci3vqOgvXEmG5LTo9wIh7zjJ2j5l+gEia4kluDDCttz+vDFFmdys5/wguHt9MMfB6ZZ7f1c+lsiCya/Zp9c7JYxNEghgrRpAmcwnRx5vNhQpFckjhbFueNtk9JAGt3rupMNXU3AsxI4cDqR31CFpGbCR93GRKIJAmxHjywbtQNznyCzr7hfey2ZJAu2D+7+//AAAA//9QSwcIgrfs6yYBAABYAQAAUEsBAhQAFAAIAAgAUXF0VXij418RAQAAHw0AAAwACQAAAAAAAAAAAAAAAAAAAG5ldHdvcmsuanNvblVUBQABSyd6Y1BLAQIUABQACAAIAFFxdFWCt+zrJgEAAFgBAAAQAAkAAAAAAAAAAAAAAFQBAABzaWduYXR1cmUuYmFzZTY0VVQFAAFLJ3pjUEsFBgAAAAACAAIAigAAAMECAAAAAA=="
	NetworkContainerEncryptedPrivateKey = "uTC09uvV2vNvwes9yj+bb80UbQh14JSFG22MSMFYgd6odNyqRw9jluBfWg4ZE38k/rugSyvchj23DN54QM3lQxuiqe35YaSvkdAsNxumOaSvsXbezY4iz/WljSfqMjKVPHtUl81RusgNZAAegp8XL7u8UBFREPSQGqhqQGggDcnp8qtXz13yo3NVFJ1Zq+kkfWT2EJyhhi+u3LIpPE4s8I/ht9LERqYeGsXbWBxCwHWHa9MWHIsa6B2naL1VEEUdpQ1GvHNY62FWFMcjpUbXNJwNDEo2GUYXYv3cbj/HFrEebhglB6FomyjkHXWEPQf7CVZx3TVjZfmCTBL5f6ud/5MOvPX2aAbeVtvLw5fT5ZocnPExAMoFlYmwfSZWRauGFAavV7FLHvtkdGDQn33Z1adkL/Bgnz68ijR9SjA+XfQn7d2OTYIgN9FhhN1m8a4wovx4geRxFxJhe0kUVDHt+gXFPavBAcNb/iGu10CDk23WkFaN+eVTRP7WKviVhUsiraTW0CsGr1E0HYy6SY0A6PbEgWEP2azg8jAZRKiYG0uLyy81JN9C55oSAXcNKooqblcX4bPB9tI1cy4zbdhj6GehfxRQ872ZsGwnpDzy3iiQstI7XYhUvQKctl7IPz4JN++sP1qUvkFaUfEeXFtzzgZ5d92qsaMqz+6yA5kFk792I+W3F/mJdQnqeLhNKN3e+PWj2UlPpjFOmcCg6d/pKHCZk6GWYHU/V37/rLYkI6uw63r6/xMmpukXcbYvH5yd/5Ej0ohrtwz3iR5ZzpXVBIOlapW50lvYoBPLpPFtIKCPj3GMp8jqut9OsiYqCCmOdTKa49regsyJtomVhU+o26i8EHrvgGqQdHhFgDY2XlUB6C/KgKuYuML/8DzTyc+0qxnxeVMwC5Ug+H8qaBJMBlj0p1TDnGIKU2B/fostvL980a6pu9OUhgbCkEvQk8ba6KbDIsVjsVV2bAGi2sx4Gsr5z6x8O0taW+pO2ZE8gjq4WNiF9iWvG4JPclux0NbNhz3CPmfxhC80is25r6KbYHm8UhCTmsa7scWxV60pArpJyyoFf65L4kvxrmEKHAvUQYXs02wQ3e/mqaJuIzkfPkJFqB0R8y68nFLf3XKZQNaBT6QKaMEBfg0/CSkHCPZMRJMu+Rn85w1n4dB+J5mRKo/0cSk+3zGbQR+NbjgNIMUb/pUkJwrKhEJppejRBuu9Yvolu+q/2rnbBEElz/OrOyqw2vP9bbv0Q0HB2JkYD97j8aPbg8Vw6SOfaoyGzbJ21LNNlfPAWF4skyi8RvEAXGRnGS/zRDkTSUqANK+unAMgRTxEFgK2b86BMQuCxm6kPAW0GdY0nqkGvrdTK6Gs+8aLSI21dPHN/n2OCoIZvOqu3l9v6mlHtONJqbvfsDFk64hztO2p0SdhVi84I19y/hMJQESkKorGVR5XH/mKCBcBanYtpKhOtQZdWNT9taImL9EKtTeXcOfjs8YPSCGunueYp+nDBSmhlrC/K1mVCBAaUXVTxe1YaVGrXRMlIYUIB2g+KkV0d7ZNxzhqNJcw1Aud322IXI9aB4c8HismCRUp7NkRmTPYlT/hBJUJ1hVCIYWvzNYXd2Cse6bksv8KSpYf1rJHszGlvlHD4T8VBs29i0XDb5UyGLmLeyYRGz4="
	NetworkContainerPublicKey           = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA8V7FEvpzVo4sLhE3rIEmKwbLmNZkweZLucv/vxIbj3y8jpJiEGT3kQA9JvGscdsS85gca34WCfKdMJBKErUm28/UAWnDZeVUmQyxwGXs2jO/OLQukwJT76Umsu/KIfr7zKxkzfm7fTsJ8q1ZYuHgndi4OTblKqy/tSynyEYFnlbpEvmIAS2ZJblarxaG5VJo3YA5ZdO5FTcuaSkZ+9v4uMvcwFK9qIigJCS+xJa+ubgN9cv2RuHuQB7+Qw9bGbCjk9cSGnbV0ttwoVMZxFkT72lAXdp5/NLWcRpKnnjEvkWKjo21ROeH6hk4qfa30Q/Q+hLbPxhLlXX2r9sNEEZkWQIDAQAB"
	NetworkContainerFileNetwork         = "network.json"
	NetworkContainerFileSignature       = "signature.base64"
)

func NetworkContainerLoadStaticDefault() (network *Network, err error) {
	var zipFileBS []byte
	zipFileBS, err = base64.StdEncoding.DecodeString(NetworkContainerDefault)
	return NetworkContainerLoad(zipFileBS, NetworkContainerPublicKey)
}

func NetworkContainerLoadFromInternet() (network *Network, err error) {
	// Load local static network
	network, err = NetworkContainerLoadStaticDefault()
	if err != nil {
		return
	}

	var httpClient *http.Client
	tr := &http.Transport{}
	httpClient = &http.Client{Transport: tr}
	httpClient.Timeout = 1 * time.Second

	networks := make([]*Network, 0)

	for _, initialPoint := range network.InitialPoints {
		var response *http.Response
		response, err = httpClient.Get(initialPoint)
		if err != nil {
			continue
		}

		var content []byte
		content, err = ioutil.ReadAll(response.Body)
		if err != nil {
			response.Body.Close()
			continue
		}
		var networkBS []byte
		networkBS, err = base64.StdEncoding.DecodeString(string(content))
		response.Body.Close()

		var n *Network
		n, err = NetworkContainerLoadDefault(networkBS)
		if err != nil {
			continue
		}
		networks = append(networks, n)
	}

	// No fresh networks - use default static network
	if len(networks) < 1 {
		return
	}

	// Get latest network
	fmt.Println("loaded networks:")
	for _, n := range networks {
		fmt.Println(n.Timestamp)
		if n.Timestamp > network.Timestamp {
			network = n
		}
	}

	return
}

func NetworkContainerLoadDefault(zipFileBS []byte) (network *Network, err error) {
	return NetworkContainerLoad(zipFileBS, NetworkContainerPublicKey)
}

func NetworkContainerLoad(zipFileBS []byte, publicKeyBase64 string) (network *Network, err error) {
	var publicKeyBS []byte
	publicKeyBS, err = base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return
	}

	var publicKeyAny any
	publicKeyAny, err = x509.ParsePKIXPublicKey(publicKeyBS)

	var publicKey *rsa.PublicKey
	var ok bool
	publicKey, ok = publicKeyAny.(*rsa.PublicKey)
	if !ok {
		err = errors.New("wrong public key")
	}

	buf := bytes.NewReader(zipFileBS)
	var zipFile *zip.Reader
	zipFile, err = zip.NewReader(buf, buf.Size())
	if err != nil {
		return
	}
	var networkBS []byte
	{
		var file fs.File
		file, err = zipFile.Open(NetworkContainerFileNetwork)
		if err == nil {
			networkBS, err = ioutil.ReadAll(file)
			if err != nil {
				_ = file.Close()
				return
			}
			_ = file.Close()
		} else {
			return
		}
	}
	var signatureBase64BS []byte
	{
		var file fs.File
		file, err = zipFile.Open(NetworkContainerFileSignature)
		if err == nil {
			signatureBase64BS, err = ioutil.ReadAll(file)
			if err != nil {
				_ = file.Close()
				return
			}
			_ = file.Close()
		} else {
			return
		}
	}

	var signature []byte
	signature, err = base64.StdEncoding.DecodeString(string(signatureBase64BS))
	if err != nil {
		return
	}

	hash := sha256.Sum256(networkBS)
	err = rsa.VerifyPSS(publicKey, crypto.SHA256, hash[:], signature, &rsa.PSSOptions{
		SaltLength: 32,
	})
	if err != nil {
		return
	}

	network, err = NewNetworkFromBytes(networkBS)
	return
}

func NetworkContainerCreateKey(privateKeyPassword string) (encryptedPrivateKeyBase64 string, publicKeyBase64 string, err error) {
	var privateKey *rsa.PrivateKey
	privateKey, err = GenerateRSAKey()
	if err != nil {
		return
	}

	var privateKeyBS []byte
	privateKeyBS, err = x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return
	}

	var publicKeyBS []byte
	publicKeyBS, err = x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return
	}

	passwordHash := sha256.Sum256([]byte(privateKeyPassword))
	aesKey := passwordHash[:]
	var encryptedPrivateKeyBS []byte
	encryptedPrivateKeyBS, err = EncryptAESGCM(privateKeyBS, aesKey)
	if err != nil {
		return
	}

	encryptedPrivateKeyBase64 = base64.StdEncoding.EncodeToString(encryptedPrivateKeyBS)
	publicKeyBase64 = base64.StdEncoding.EncodeToString(publicKeyBS)

	return
}

func NetworkContainerMake(network *Network, encryptedPrivateKeyBase64 string, privateKeyPassword string) (resultZipFile []byte, err error) {
	if network == nil {
		err = errors.New("network == nil")
		return
	}

	var encryptedPrivateKeyBS []byte
	encryptedPrivateKeyBS, err = base64.StdEncoding.DecodeString(encryptedPrivateKeyBase64)
	if err != nil {
		return
	}

	passwordHash := sha256.Sum256([]byte(privateKeyPassword))
	aesKey := passwordHash[:]
	var privateKeyBS []byte
	privateKeyBS, err = DecryptAESGCM(encryptedPrivateKeyBS, aesKey)
	if err != nil {
		return
	}

	var privateKeyAny any
	privateKeyAny, err = x509.ParsePKCS8PrivateKey(privateKeyBS)
	if err != nil {
		return
	}

	var privateKey *rsa.PrivateKey
	var ok bool
	privateKey, ok = privateKeyAny.(*rsa.PrivateKey)
	if !ok {
		err = errors.New("wrong private key")
		return
	}

	if privateKey == nil {
		err = errors.New("privateKey == nil")
		return
	}

	networkBS := network.toBytes()
	hash := sha256.Sum256(networkBS)
	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hash[:], &rsa.PSSOptions{
		SaltLength: 32,
	})
	if err != nil {
		return
	}

	signatureBase64 := base64.StdEncoding.EncodeToString(signature)

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	{
		var zipFile io.Writer

		header := &zip.FileHeader{
			Name:     NetworkContainerFileNetwork,
			Method:   zip.Deflate,
			Modified: time.Now(),
		}
		zipFile, err = zipWriter.CreateHeader(header)
		if err == nil {
			_, err = zipFile.Write(networkBS)
			if err != nil {
				zipWriter.Close()
				return
			}
		} else {
			zipWriter.Close()
			return
		}
	}
	{
		var zipFile io.Writer
		header := &zip.FileHeader{
			Name:     NetworkContainerFileSignature,
			Method:   zip.Deflate,
			Modified: time.Now(),
		}
		zipFile, err = zipWriter.CreateHeader(header)
		if err == nil {
			_, err = zipFile.Write([]byte(signatureBase64))
			if err != nil {
				zipWriter.Close()
				return
			}
		} else {
			zipWriter.Close()
			return
		}
	}
	err = zipWriter.Close()
	resultZipFile = buf.Bytes()
	return
}
