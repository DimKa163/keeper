package auth

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_Authenticate_Success(t *testing.T) {
	sut := createAuthService()
	password := "password"
	password2 := "password"

	pwd, salt, err := sut.GenerateHash([]byte(password))
	if err != nil {
		t.Fatal(err)
	}

	err = sut.Authenticate([]byte(password2), pwd, salt)

	assert.NoError(t, err)
}

func TestAuthService_Authenticate_Fail(t *testing.T) {
	sut := createAuthService()
	password := "password"
	password2 := "dorwssap"

	pwd, salt, err := sut.GenerateHash([]byte(password))
	if err != nil {
		t.Fatal(err)
	}

	err = sut.Authenticate([]byte(password2), pwd, salt)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPassword)
}

func TestGenerateSalt(t *testing.T) {
	sut := createAuthService()

	salt1, err := sut.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt returned error: %v", err)
	}

	salt2, _ := sut.GenerateSalt()
	if bytes.Equal(salt1, salt2) {
		t.Error("expected salts to be different, actually identical")
	}
}
func createAuthService() *authService {
	return &authService{&ArgonConfig{
		SaltLength:  16,
		Iterations:  4,
		Parallelism: 2,
		Memory:      64 * 1024,
		KeyLength:   32,
	}}
}
