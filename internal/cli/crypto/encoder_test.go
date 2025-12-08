package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAesEncoder_Encode(t *testing.T) {
	key := generateRandomKey()
	plaintext := []byte("hello world")
	decoder := NewAesDecoder()
	encoder := NewAesEncoder()
	ciphertext, err := encoder.Encode(plaintext, key)
	assert.NoError(t, err)
	plaintext2, err := decoder.Decode(ciphertext, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, plaintext2)
}
