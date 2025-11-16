package core

import (
	"encoding/json"
	"time"
)

type Conflict struct {
	ID         int32
	CreatedAt  time.Time
	ModifiedAt time.Time
	RecordID   string
	Local      *ConflictItem
	Remote     *ConflictItem
}

type ConflictItem struct {
	Record  *Record `json:"record"`
	Deleted bool
}

func (c *Conflict) MarshalLocal() ([]byte, error) {
	return json.Marshal(c.Local)
}

func (c *Conflict) MarshalRemote() ([]byte, error) {
	return json.Marshal(c.Remote)
}
