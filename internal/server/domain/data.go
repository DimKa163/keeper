package domain

import (
	"context"
	"errors"
	"time"

	"github.com/beevik/guid"
)

var (
	ErrDataConflict = errors.New("data conflict")
)

type DataType int

const (
	LoginPassType DataType = iota
	TextType
	BankCardType
	OtherType
)

func (d DataType) String() string {
	return [...]string{"login_pass", "text", "bank_card", "other"}[d]
}

type Data struct {
	ID           guid.Guid
	CreatedAt    time.Time
	Name         string
	UserID       guid.Guid
	Type         DataType
	Large        bool
	DekNonce     []byte
	Dek          []byte
	PayloadNonce []byte
	Payload      []byte
	Version      int32
}

func (sd *Data) Update(name string, large bool, dekNonce, dek, dataNonce, data []byte, version int32) error {
	if version != sd.Version {
		return ErrDataConflict
	}
	sd.Name = name
	sd.Large = large
	sd.DekNonce = dekNonce
	sd.Dek = dek
	sd.PayloadNonce = dataNonce
	sd.Payload = data
	return nil
}

func (sd *Data) UpVersion() {
	sd.Version += 1
}

type FilePart struct {
	ID     guid.Guid
	DataID int64
	Path   string
	Nonce  []byte
}

type DataRepository interface {
	Get(ctx context.Context, id guid.Guid) (*Data, error)
	GetAll(ctx context.Context, userID guid.Guid, limit, skip int) ([]*Data, error)
	Insert(ctx context.Context, data *Data) error
	Update(ctx context.Context, data *Data) error
	Delete(ctx context.Context, id guid.Guid) error
}

type FilePartRepository interface {
	Get(ctx context.Context, dataID int64) ([]*FilePart, error)
}

type IDataProvider interface {
	ExecuteWriter(ctx context.Context, dataID guid.Guid) Writer

	ExecuteReader(ctx context.Context, dataID guid.Guid) Reader
}

type Writer interface {
	Write(ctx context.Context, part *FilePart) error
}

type Reader interface {
	Read(ctx context.Context) (*FilePart, error)
}
