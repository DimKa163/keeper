package core

type Encoder interface {
	Encode(data, key []byte) ([]byte, error)
}
