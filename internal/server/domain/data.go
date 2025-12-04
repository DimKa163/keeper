package domain

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/beevik/guid"
)

var (
	ErrDataConflict = errors.New("data conflict")
)

type OperationType int

const (
	InsertOperation OperationType = iota
	UpdateOperation
	DeleteOperation
)

func (ot OperationType) String() string {
	return [...]string{"insert", "update", "delete"}[ot]
}

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
	ID         guid.Guid
	CreatedAt  time.Time
	ModifiedAt time.Time
	UserID     guid.Guid
	Type       DataType
	BigData    bool
	Dek        []byte
	Payload    []byte
	Path       string
	Version    int32
	Deleted    bool
}

func (sd *Data) Update(large bool, dek, data []byte, deleted bool, version int32) {
	sd.ModifiedAt = time.Now()
	sd.BigData = large
	sd.Dek = dek
	sd.Payload = data
	sd.Deleted = deleted
	sd.Version = version
}

func (sd *Data) File() (*os.File, error) {
	return os.Open(fmt.Sprintf("%s_%d", sd.Path, sd.Version))
}

type DataRepository interface {
	Get(ctx context.Context, id guid.Guid) (*Data, error)
	GetAll(ctx context.Context, userID guid.Guid, greaterThan int32) ([]*Data, error)
	Insert(ctx context.Context, data *Data) error
	Update(ctx context.Context, data *Data) error
	Delete(ctx context.Context, data *Data) error
}

type DataService interface {
	PushUnary(ctx context.Context, data *Data) error

	PushMetadata(ctx context.Context, data *Data) error

	PushData(ctx context.Context, id guid.Guid, data []byte, version int32) error

	Finish(ctx context.Context, id guid.Guid, version int32) error

	OpenFile(fileName string) (io.ReadCloser, error)

	Push(ctx context.Context, data []*Data) error

	Poll(ctx context.Context, since int32) ([]*Data, int32, error)
}
