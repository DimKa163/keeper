package crypto

import (
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

func (a *AesEncoder) Encode(record *core.Record, data, materKey []byte) error {
	dek, err := generateDek(a.dekLength)
	if err != nil {
		return err
	}
	record.DataNonce, record.Data, err = seal(data, dek)
	if err != nil {
		return err
	}
	record.DekNonce, record.Dek, err = seal(dek, materKey)
	if err != nil {
		return err
	}
	return nil
}

func generateDek(len int32) ([]byte, error) {
	dek := make([]byte, len)
	if _, err := rand.Read(dek); err != nil {
		return nil, err
	}
	return dek, nil
}

func seal(data []byte, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	cipherData := gcm.Seal(nil, nonce, data, nil)
	return nonce, cipherData, nil
}
