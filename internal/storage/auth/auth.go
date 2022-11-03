package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const decodedTokenLen = 32

// GenerateRandomToken produces a URL safe base64 encoded string of random characters
// the data is sourced from a pseudo-random input stream
func GenerateRandomToken() string {
	var token [decodedTokenLen]byte
	if _, err := rand.Read(token[:]); err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(token[:])
}

// HashClientToken performs a SHA256 sum on the input string
// it returns the result as a URL safe base64 encoded string
func HashClientToken(token string) (string, error) {
	// produce SHA256 hash of token
	hash := sha256.New()
	if _, err := hash.Write([]byte(token)); err != nil {
		return "", fmt.Errorf("hashing client token: %w", err)
	}

	// base64(sha256sum)
	var (
		data = make([]byte, 0, base64.URLEncoding.EncodedLen(hash.Size()))
		buf  = bytes.NewBuffer(data)
		enc  = base64.NewEncoder(base64.URLEncoding, buf)
	)

	if _, err := enc.Write(hash.Sum(nil)); err != nil {
		return "", fmt.Errorf("hashing client token: %w", err)
	}

	if err := enc.Close(); err != nil {
		return "", fmt.Errorf("hashing client token: %w", err)
	}

	return buf.String(), nil
}
