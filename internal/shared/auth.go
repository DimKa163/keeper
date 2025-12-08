// Package shared common tools
package shared

import (
	"crypto/rand"

	"golang.org/x/crypto/argon2"
)

func Hash(pwd []byte, salt []byte, time, memory, keyLen uint32, threads uint8) []byte {
	return argon2.IDKey(pwd, salt, time, memory, threads, keyLen)
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	return salt, err
}

func Compare(b, c []byte) bool {
	if len(b) != len(c) {
		return false
	}
	var result byte
	for i := range b {
		result |= b[i] ^ c[i]
	}
	return result == 0
}
