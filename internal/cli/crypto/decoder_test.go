package crypto

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAesDecoder_Decode(t *testing.T) {
	key := generateRandomKey()
	plainText := []byte("Hello, world!")

	block, err := aes.NewCipher(key)
	assert.NoError(t, err)

	gcm, err := cipher.NewGCM(block)
	assert.NoError(t, err)

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	assert.NoError(t, err)

	cipherData := gcm.Seal(nonce, nonce, plainText, nil)

	decoder := NewAesDecoder()
	decryptedData, err := decoder.Decode(cipherData, key)
	assert.NoError(t, err)
	assert.Equal(t, plainText, decryptedData)
}

func TestGzipDecoder_Decode(t *testing.T) {
	plainText := []byte("This is some text to compress.")

	var compressedData bytes.Buffer
	writer := gzip.NewWriter(&compressedData)
	_, err := writer.Write(plainText)
	assert.NoError(t, err)
	writer.Close()

	key := generateRandomKey()
	decoder := NewAesDecoder()

	block, err := aes.NewCipher(key)
	assert.NoError(t, err)

	gcm, err := cipher.NewGCM(block)
	assert.NoError(t, err)

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	assert.NoError(t, err)

	cipherData := gcm.Seal(nonce, nonce, compressedData.Bytes(), nil)

	gzipDecoder := NewGzipDecoder(decoder)
	decryptedData, err := gzipDecoder.Decode(cipherData, key)
	assert.NoError(t, err)

	assert.Equal(t, plainText, decryptedData)
}

func TestAesDecoder_Decode_InvalidKey(t *testing.T) {
	plainText := []byte("Hello, world!")
	key := generateRandomKey()

	block, err := aes.NewCipher(key)
	assert.NoError(t, err)

	gcm, err := cipher.NewGCM(block)
	assert.NoError(t, err)

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	assert.NoError(t, err)

	cipherData := gcm.Seal(nonce, nonce, plainText, nil)

	invalidKey := generateRandomKey()

	decoder := NewAesDecoder()
	_, err = decoder.Decode(cipherData, invalidKey)
	assert.Error(t, err)
}

func generateRandomKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	return key
}
