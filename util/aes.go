package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// SimpleAES 包装aes，使得使用更简单点
type SimpleAES struct {
	gcm cipher.AEAD
}

// NewSimpleAES 构造SimpleAES
func NewSimpleAES(key []byte) (*SimpleAES, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &SimpleAES{
		gcm: aesgcm,
	}, nil
}

// Enc 加密data，返回的string包含nonce和加密后的data
func (s *SimpleAES) Enc(data []byte) (string, error) {
	nonceSize := s.gcm.NonceSize()
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherData := s.gcm.Seal(nil, nonce, data, nil)
	bs := append(nonce[:], cipherData[:]...)

	return base64.URLEncoding.EncodeToString(bs), nil
}

// Dec 解密data，返回当初传入Enc的数据
func (s *SimpleAES) Dec(data string) ([]byte, error) {
	bs, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	nonceSize := s.gcm.NonceSize()
	return s.gcm.Open(nil, bs[:nonceSize], bs[nonceSize:], nil)
}
