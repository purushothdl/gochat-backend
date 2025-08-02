package tokenutil

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Generate creates a secure, random token and its corresponding SHA-256 hash.
func Generate() (plaintext string, hash string, err error) {
	// Generate a 32-byte random token.
	p := make([]byte, 32)
	if _, err := rand.Read(p); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	plaintext = hex.EncodeToString(p)

	// Hash the plaintext token.
	hashBytes := sha256.Sum256([]byte(plaintext))
	hash = hex.EncodeToString(hashBytes[:])

	return plaintext, hash, nil
}

// Hash generates the SHA-256 hash of a plaintext token.
func Hash(plaintext string) string {
	hashBytes := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(hashBytes[:])
}