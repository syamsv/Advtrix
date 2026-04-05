package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

var gcm cipher.AEAD

// Init initialises AES-256-GCM with the provided hex-encoded 32-byte key.
func Init(hexKey string) {
	key, err := hex.DecodeString(hexKey)
	if err != nil || len(key) != 32 {
		panic("ENCRYPTION_KEY must be a 64-character hex string (32 bytes)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("failed to create AES cipher: " + err.Error())
	}

	gcm, err = cipher.NewGCM(block)
	if err != nil {
		panic("failed to create GCM: " + err.Error())
	}
}

// Encrypt encrypts plaintext and returns a hex-encoded nonce+ciphertext.
func Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decodes hex and decrypts nonce+ciphertext back to plaintext.
func Decrypt(encoded string) (string, error) {
	data, err := hex.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
