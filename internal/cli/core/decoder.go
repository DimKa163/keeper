package core

type Decoder interface {
	Decode(data, key []byte) ([]byte, error)
}
