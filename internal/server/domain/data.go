package domain

import (
	"context"
	"errors"
	"time"

	"github.com/beevik/guid"
)

type StoredDataType int

var (
	ErrDataConflict = errors.New("data conflict")
)

const (
	LoginPassType StoredDataType = iota
	TextType
	BankCardType
	OtherType
)

func (d StoredDataType) String() string {
	return [...]string{"login_pass", "text", "bank_card", "other"}[d]
}

type StoredData struct {
	ID        guid.Guid
	CreatedAt time.Time
	Name      string
	UserID    guid.Guid
	Type      StoredDataType
	Large     bool
	DekNonce  []byte
	Dek       []byte
	DataNonce []byte
	Data      []byte
	Version   int32
}

func (sd *StoredData) Update(name string, large bool, dekNonce, dek, dataNonce, data []byte, version int32) error {
	if version <= sd.Version {
		return ErrDataConflict
	}
	sd.Name = name
	sd.Large = large
	sd.DekNonce = dekNonce
	sd.Dek = dek
	sd.DataNonce = dataNonce
	sd.Data = data
	return nil
}

func (sd *StoredData) UpVersion() {
	sd.Version += 1
}

type FilePart struct {
	ID     guid.Guid
	DataID int64
	Path   string
	Nonce  []byte
}

type StoredDataRepository interface {
	Get(ctx context.Context, id guid.Guid) (*StoredData, error)
	GetAll(ctx context.Context, userID guid.Guid, limit, skip int) ([]*StoredData, error)
	Insert(ctx context.Context, data *StoredData) error
	Update(ctx context.Context, data *StoredData) error
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
