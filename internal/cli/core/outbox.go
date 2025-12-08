package core

import (
	"encoding/json"
	"time"
)

type OperationType int

const (
	Added OperationType = iota
	Modified
	Deleted
)

type Outbox struct {
	ID        int64
	CreatedAt time.Time
	Message   []byte
	Type      OperationType
}

func NewOutbox(record *Record, tp OperationType) (*Outbox, error) {
	msg, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	return &Outbox{
		CreatedAt: time.Now(),
		Message:   msg,
		Type:      tp,
	}, nil
}

func (o *Outbox) Record() (*Record, error) {
	var record Record
	if err := json.Unmarshal(o.Message, &record); err != nil {
		return nil, err
	}
	return &record, nil
}
