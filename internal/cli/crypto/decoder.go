// Package crypto tools to crypt/encrypt data
package crypto

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"io"

	"github.com/DimKa163/keeper/internal/cli/core"
)

type AesDecoder struct {
}

func NewAesDecoder() core.Decoder {
	return &AesDecoder{}
}

func (a *AesDecoder) Decode(cipherData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, data := cipherData[:nonceSize], cipherData[nonceSize:]
	return gcm.Open(nil, nonce, data, nil)
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

type GzipDecoder struct {
	decoder core.Decoder
}

func NewGzipDecoder(decoder core.Decoder) core.Decoder {
	return &GzipDecoder{
		decoder: decoder,
	}
}

func (g GzipDecoder) Decode(data, key []byte) ([]byte, error) {
	data, err := g.decoder.Decode(data, key)
	if err != nil {
		return nil, err
	}
	reader, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

type FileDecoder struct {
	decoder core.Decoder
	fs      io.ReadCloser
	dek     []byte
}

func (f *FileDecoder) Read(p []byte) (n int, err error) {
	data, err := io.ReadAll(f.fs)
	if err != nil {
		return 0, err
	}
	d, err := f.decoder.Decode(data, f.dek)
	if err != nil {
		return -1, err
	}
	n = copy(p, d)
	return n, nil
}

func (f *FileDecoder) Close() error {
	return f.fs.Close()
}

func NewFileDecoder(decoder core.Decoder, fs io.ReadCloser, dek []byte) io.ReadCloser {
	return &FileDecoder{
		decoder: decoder,
		fs:      fs,
		dek:     dek,
	}
}
