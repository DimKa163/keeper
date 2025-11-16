package core

import (
	"encoding/json"
	"time"

	"github.com/beevik/guid"
)

type DataType int

const (
	LoginPassType DataType = iota
	TextType
	BankCardType
	OtherType
)

type Record struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
	Type       DataType  `json:"type"`
	Data       []byte    `json:"data"`
	DataNonce  []byte    `json:"data_nonce"`
	Dek        []byte    `json:"dek"`
	DekNonce   []byte    `json:"dek_nonce"`
	Version    int32     `json:"version"`
}

func CreateRecord(tp DataType) *Record {
	return &Record{
		ID:        guid.NewString(),
		CreatedAt: time.Now(),
		Type:      tp,
		Version:   1,
	}
}

func (r *Record) Encode(encoder Encoder, data, masterKey []byte) error {
	return encoder.Encode(r, data, masterKey)
}

func (r *Record) Decode(decoder Decoder, masterKey []byte) ([]byte, error) {
	data, err := decoder.Decode(r, masterKey)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Record) DecodeLoginPass(decoder Decoder, masterKey []byte) (*LoginPass, error) {
	data, err := r.Decode(decoder, masterKey)
	if err != nil {
		return nil, err
	}
	var loginPass LoginPass
	if err = json.Unmarshal(data, &loginPass); err != nil {
		return nil, err
	}
	return &loginPass, nil
}

func (r *Record) DecodeText(decoder Decoder, masterKey []byte) (*Text, error) {
	data, err := r.Decode(decoder, masterKey)
	if err != nil {
		return nil, err
	}
	var text Text
	if err = json.Unmarshal(data, &text); err != nil {
		return nil, err
	}
	return &text, nil
}

func (r *Record) DecodeBinary(decoder Decoder, masterKey []byte) (*Binary, error) {
	data, err := r.Decode(decoder, masterKey)
	if err != nil {
		return nil, err
	}
	var binary Binary
	if err = json.Unmarshal(data, &binary); err != nil {
		return nil, err
	}
	return &binary, nil
}

func (r *Record) DecodeBankCard(decoder Decoder, masterKey []byte) (*BankCard, error) {
	data, err := r.Decode(decoder, masterKey)
	if err != nil {
		return nil, err
	}
	var card BankCard
	if err = json.Unmarshal(data, &card); err != nil {
		return nil, err
	}
	return &card, nil
}
