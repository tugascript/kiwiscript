package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/argon2"
)

const memory uint32 = 65_536
const iterations uint32 = 3
const parallelism uint8 = 4
const saltSize uint32 = 16
const keySize uint32 = 32

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func HashPassword(password string) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keySize)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	return b64Salt + "." + b64Hash, nil
}

func VerifyPassword(password, hash string) bool {
	parts := strings.Split(hash, ".")

	if len(parts) != 2 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	comparisonHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keySize)
	return bytes.Equal(decodedHash, comparisonHash)
}
