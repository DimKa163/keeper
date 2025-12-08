package core

import (
	"encoding/json"
	"time"

	"github.com/DimKa163/keeper/internal/shared"

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
	BigData    bool      `json:"big_data"`
	Data       []byte    `json:"data"`
	Dek        []byte    `json:"dek"`
	Version    int32     `json:"version"`
	Deleted    bool      `json:"deleted"`
	Corrupted  bool      `json:"corrupted"`
}

func CreateRecord(tp DataType) *Record {
	return &Record{
		ID:        guid.NewString(),
		CreatedAt: time.Now(),
		Type:      tp,
	}
}

//func (r *Record) Encode(encoder Encoder, data, masterKey []byte) error {
//	return encoder.Encode(r, data, masterKey)
//}

func (r *Record) IsChanged(state *SyncState) bool {
	return r.Version > state.Value
}
func (r *Record) Decode(decoder Decoder, masterKey []byte) ([]byte, error) {
	dek, err := decoder.Decode(r.Dek, masterKey)
	if err != nil {
		return nil, err
	}
	data, err := decoder.Decode(r.Data, dek)
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

func (r *Record) Validate(fileProvider *shared.FileProvider) error {
	return nil
}
