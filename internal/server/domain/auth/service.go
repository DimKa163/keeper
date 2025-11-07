package auth

import (
	"crypto/rand"
	"errors"

	"golang.org/x/crypto/argon2"
)

var ErrInvalidPassword = errors.New("invalid password")

type ArgonConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint32
	SaltLength  uint32
	KeyLength   uint32
}

// AuthService authenticate users
type AuthService interface {
	// GenerateHash hashes the password and generate salt
	GenerateHash(password []byte) (pwd, salt []byte, err error)
	// Authenticate compare typed password and hashed password
	Authenticate(password, hashedPassword, salt []byte) error

	GenerateSalt() ([]byte, error)
}

type authService struct {
	*ArgonConfig
}

func NewAuthService(config *ArgonConfig) AuthService {
	return &authService{ArgonConfig: config}
}

func (a *authService) GenerateHash(password []byte) (pwd, salt []byte, err error) {
	salt, err = a.GenerateSalt()
	if err != nil {
		pwd = nil
		salt = nil
		return
	}
	pwd = a.hash(password, salt)
	return
}

func (a *authService) GenerateSalt() ([]byte, error) {
	salt := make([]byte, a.SaltLength)
	_, err := rand.Read(salt)
	return salt, err
}
func (a *authService) Authenticate(password, hashedPassword, salt []byte) error {
	candidateHash := a.hash(password, salt)
	if !a.compare(hashedPassword, candidateHash) {
		return ErrInvalidPassword
	}
	return nil
}

func (a *authService) hash(pwd []byte, salt []byte) []byte {
	return argon2.IDKey(pwd, salt, a.Iterations, a.Memory, uint8(a.Parallelism), a.KeyLength)
}

func (a *authService) compare(b, c []byte) bool {
	if len(b) != len(c) {
		return false
	}
	var result byte
	for i := range b {
		result |= b[i] ^ c[i]
	}
	return result == 0
}
