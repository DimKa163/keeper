package core

type Decoder interface {
	Decode(nonce, data, key []byte) ([]byte, error)
}
