package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"wsf/config"
)

// Cipher handles encryption
type Cipher struct {
	key []byte
}

// EncodeString encrypts string
func (c *Cipher) EncodeString(s string) (string, error) {
	encoding := []byte(s)
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(encoding))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], encoding)
	return string(ciphertext), nil
}

// DecodeString decrypts string
func (c *Cipher) DecodeString(s string) (string, error) {
	encoding := []byte(s)
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(encoding))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	result := make([]byte, len(encoding))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(result, ciphertext[aes.BlockSize:])

	return string(result), nil
}

// Encode encrypts bytes array
func (c *Cipher) Encode(encoding []byte, useiv bool) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(encoding))
	var iv []byte
	if useiv {
		iv = ciphertext[:aes.BlockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, err
		}
	} else {
		iv = make([]byte, 0)
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], encoding)
	return ciphertext, nil
}

// Decode decrypts bytes array
func (c *Cipher) Decode(encoding []byte, useiv bool) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(encoding))
	var iv []byte
	if useiv {
		iv = ciphertext[:aes.BlockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, err
		}
	} else {
		iv = make([]byte, 0)
	}

	result := make([]byte, len(encoding))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(result, ciphertext[aes.BlockSize:])

	return result, nil
}

// NewCipher creates a new cipher
func NewCipher(key []byte) (*Cipher, error) {
	if key == nil {
		k := config.App.GetString("application.CryptoKey")
		if k == "" {
			return nil, errors.New("[Crypt] CryptoKey is undefiend")
		}

		key = []byte(k)
	}

	if len(key) < 16 {
		key = append(key, make([]byte, 16-len(key))...)
	} else if len(key) > 16 && len(key) < 32 {
		key = append(key, make([]byte, 32-len(key))...)
	} else if len(key) > 32 && len(key) < 64 {
		key = append(key, make([]byte, 64-len(key))...)
	} else if len(key) > 64 {
		key = append([]byte{}, key[:64]...)
	}

	return &Cipher{
		key: key,
	}, nil
}

// Encode encrypts bytes array
func Encode(encoding []byte, key []byte, useiv bool) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(encoding))
	var iv []byte
	if useiv {
		iv = ciphertext[:aes.BlockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, err
		}
	} else {
		iv = make([]byte, 0)
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], encoding)
	return ciphertext, nil
}

// Decode decrypts bytes array
func Decode(encoding []byte, key []byte, useiv bool) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(encoding))
	var iv []byte
	if useiv {
		iv = ciphertext[:aes.BlockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, err
		}
	} else {
		iv = make([]byte, 0)
	}

	result := make([]byte, len(encoding))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(result, ciphertext[aes.BlockSize:])

	return result, nil
}
