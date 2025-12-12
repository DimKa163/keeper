package security

import (
	"testing"
	"time"

	"github.com/beevik/guid"

	"github.com/stretchr/testify/assert"
)

func TestJWTEngine_GenerateToken(t *testing.T) {
	config := &JWTConfig{
		TokenExpiration: time.Minute,
		SecretKey:       []byte("secret"),
	}
	sut := createJWTEngine(config)

	token, err := sut.GenerateToken(*guid.New())
	if err != nil {
		t.Fatalf("generate tokent return err: %v", err)
	}

	if token == "" {
		t.Error("token is empty")
	}
}

func TestJWTEngine_GenerateToken_WithEmptyKeyShouldBeError(t *testing.T) {
	config := &JWTConfig{
		TokenExpiration: time.Minute,
		SecretKey:       nil,
	}
	sut := createJWTEngine(config)

	_, err := sut.GenerateToken(*guid.New())
	assert.Error(t, err)
}

func createJWTEngine(config *JWTConfig) *JWTEngine {
	return &JWTEngine{config}
}
