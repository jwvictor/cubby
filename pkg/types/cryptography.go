package types

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"golang.org/x/crypto/scrypt"
	"io"
)

const (
	SaltString = "cbbc"
)

func EncryptSymmetric(body, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("InvalidKeyLength")
	}
	// generate a new aes cipher using our 32 byte long key
	c, err := aes.NewCipher(key)
	// if there are any errors, handle them
	if err != nil {
		return nil, err
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//gcm, err := cipher.NewGCM(c)

	//iv := make([]byte, aes.BlockSize)
	//if _, err := rand.Read(iv); err != nil {
	//	return nil, err
	//}

	iv := make([]byte, aes.BlockSize)
	// populates our IV with a cryptographically secure
	// random sequence
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewOFB(c, iv)
	out := make([]byte, len(body))
	stream.XORKeyStream(out, body)

	var result []byte = iv
	result = append(result, out...)

	return result, nil
}

func DecryptSymmetric(encdata, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := encdata[:aes.BlockSize]
	ciphertext := encdata[aes.BlockSize:]

	stream := cipher.NewOFB(c, iv)

	out := make([]byte, len(ciphertext))
	stream.XORKeyStream(out, ciphertext)
	return out, nil
}

func DeriveSymmetricKey(key string) ([]byte, error) {
	rig, err := scrypt.Key([]byte(key), []byte(SaltString), 16384, 8, 1, 32)
	return rig, err
}
