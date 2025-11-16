package crypto

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/DimKa163/keeper/internal/cli/core"
)

type AesDecoder struct {
}

func NewAesDecoder() core.Decoder {
	return &AesDecoder{}
}

func (a *AesDecoder) Decode(record *core.Record, masterKey []byte) ([]byte, error) {
	dek, err := a.aesOpen(masterKey, record.DekNonce, record.Dek)
	if err != nil {
		return nil, err
	}
	data, err := a.aesOpen(dek, record.DataNonce, record.Data)
	return data, err
}

func (a *AesDecoder) aesOpen(key, nonce []byte, cipherData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	data, err := gcm.Open(nil, nonce, cipherData, nil)
	return data, err
}
