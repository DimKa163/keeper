package core

type Encoder interface {
	Encode(record *Record, data, materKey []byte) error
}
