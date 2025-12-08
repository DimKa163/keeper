package crypto

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"github.com/DimKa163/keeper/internal/cli/core"
)

type AesEncoder struct {
	dekLength int32
}

func NewAesEncoder() core.Encoder {
	return &AesEncoder{
		dekLength: 32,
	}
}

func (a *AesEncoder) Encode(data, key []byte) (cipherData []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	cipherData = gcm.Seal(nonce, nonce, data, nil)
	return cipherData, nil
}

type GzipEncoder struct {
	encoder core.Encoder
}

func NewGzipEncoder(encoder core.Encoder) core.Encoder {
	return &GzipEncoder{
		encoder: encoder,
	}
}

func (g *GzipEncoder) Encode(data, key []byte) (cipherData []byte, err error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
	if err != nil {
		return
	}
	_, err = gz.Write(data)
	if err != nil {
		return
	}
	if err = gz.Close(); err != nil {
		return nil, err
	}
	return g.encoder.Encode(buf.Bytes(), key)
}
