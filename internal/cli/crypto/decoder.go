package crypto

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"github.com/DimKa163/keeper/internal/cli/core"
	"io"
)

type AesDecoder struct {
}

func NewAesDecoder() core.Decoder {
	return &AesDecoder{}
}

func (a *AesDecoder) Decode(nonce, data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
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

func (g GzipDecoder) Decode(nonce, data, key []byte) ([]byte, error) {
	data, err := g.decoder.Decode(nonce, data, key)
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
	nonce   []byte
}

func (f *FileDecoder) Read(p []byte) (n int, err error) {
	data, err := io.ReadAll(f.fs)
	if err != nil {
		return 0, err
	}
	d, err := f.decoder.Decode(f.nonce, data, f.dek)
	if err != nil {
		return -1, err
	}
	n = copy(p, d)
	return n, nil
}

func (f *FileDecoder) Close() error {
	return f.fs.Close()
}

func NewFileDecoder(decoder core.Decoder, fs io.ReadCloser, dek, nonce []byte) io.ReadCloser {
	return &FileDecoder{
		decoder: decoder,
		fs:      fs,
		dek:     dek,
		nonce:   nonce,
	}
}
