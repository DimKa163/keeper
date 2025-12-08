package auth

import (
	"crypto/rand"
	"errors"

	shared "github.com/DimKa163/keeper/internal/shared"
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
	salt, err = shared.GenerateSalt()
	if err != nil {
		pwd = nil
		salt = nil
		return
	}
	pwd = shared.Hash(password, salt, a.Iterations, a.Memory, a.KeyLength, uint8(a.Parallelism))
	return
}

func (a *authService) GenerateSalt() ([]byte, error) {
	salt := make([]byte, a.SaltLength)
	_, err := rand.Read(salt)
	return salt, err
}
func (a *authService) Authenticate(password, hashedPassword, salt []byte) error {
	candidateHash := shared.Hash(password, salt, a.Iterations, a.Memory, a.KeyLength, uint8(a.Parallelism))
	if !shared.Compare(hashedPassword, candidateHash) {
		return ErrInvalidPassword
	}
	return nil
}
