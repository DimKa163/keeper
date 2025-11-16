package core

type Decoder interface {
	Decode(record *Record, masterKey []byte) ([]byte, error)
}
