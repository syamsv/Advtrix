package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/syamsv/Advtrix/config"
	"github.com/syamsv/Advtrix/common/nts"
)

var (
	ErrEmptySecret   = errors.New("totp: secret must not be empty")
	ErrInvalidSecret = errors.New("totp: secret is not valid base32")
	ErrEmptyCode     = errors.New("totp: code must not be empty")
	ErrInvalidCode   = errors.New("totp: code must be exactly 6 digits")
)

// GenerateSecret creates a cryptographically random base32-encoded secret
// of the given byte length (recommended: 20 for 160-bit key).
func GenerateSecret(length int) (string, error) {
	if length <= 0 {
		length = 20
	}

	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("totp: failed to generate random bytes: %w", err)
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), nil
}

// GenerateTOTP generates a 6-digit TOTP code using NTS-synced atomic time
// with a configurable step interval. Returns the code, seconds remaining
// in the current step, and any error.
func GenerateTOTP(secret string) (string, int, error) {
	key, err := decodeSecret(secret)
	if err != nil {
		return "", 0, err
	}

	now := nts.Now()
	step := config.STEPSECOND
	counter := uint64(now.Unix()) / uint64(step)

	hash := calculateHash(key, counter)
	code := truncate(hash)
	remaining := step - (int(now.Unix()) % step)

	return fmt.Sprintf("%06d", code), remaining, nil
}

// ValidateTOTP checks a code against the current and immediately adjacent
// time steps (±1) to account for minor transmission delay.
func ValidateTOTP(secret, code string) (bool, error) {
	if code == "" {
		return false, ErrEmptyCode
	}
	if len(code) != 6 {
		return false, ErrInvalidCode
	}

	key, err := decodeSecret(secret)
	if err != nil {
		return false, err
	}

	now := nts.Now()
	currentCounter := uint64(now.Unix()) / uint64(config.STEPSECOND)

	for _, counter := range []uint64{currentCounter - 1, currentCounter, currentCounter + 1} {
		hash := calculateHash(key, counter)
		expected := fmt.Sprintf("%06d", truncate(hash))
		if hmac.Equal([]byte(expected), []byte(code)) {
			return true, nil
		}
	}

	return false, nil
}

// decodeSecret normalises and decodes a base32 secret into raw bytes.
func decodeSecret(secret string) ([]byte, error) {
	if secret == "" {
		return nil, ErrEmptySecret
	}

	secret = strings.ReplaceAll(secret, " ", "")
	secret = strings.TrimRight(secret, "=")
	secret = strings.ToUpper(secret)

	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return nil, ErrInvalidSecret
	}

	return key, nil
}

func calculateHash(key []byte, counter uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	h := hmac.New(sha1.New, key)
	h.Write(buf)

	return h.Sum(nil)
}

func truncate(hash []byte) int {
	offset := hash[len(hash)-1] & 0x0f

	code := int(hash[offset]&0x7f)<<24 |
		int(hash[offset+1])<<16 |
		int(hash[offset+2])<<8 |
		int(hash[offset+3])

	return code % 1000000
}
