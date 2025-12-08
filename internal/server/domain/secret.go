package domain

import (
	"context"
	"time"

	"github.com/beevik/guid"
)

type SecretType int

const (
	LoginPassType SecretType = iota
	TextType
	BankCardType
	OtherType
)

func (d SecretType) String() string {
	return [...]string{"login_pass", "text", "bank_card", "other"}[d]
}

type Secret struct {
	ID         guid.Guid
	CreatedAt  time.Time
	ModifiedAt time.Time
	UserID     guid.Guid
	Type       SecretType
	BigData    bool
	Dek        []byte
	Payload    []byte
	Path       string
	Version    int32
	Deleted    bool
}
type SecretRepository interface {
	Get(ctx context.Context, id guid.Guid) (*Secret, error)
	GetAll(ctx context.Context, userID guid.Guid, greaterThan int32) ([]*Secret, error)
	Insert(ctx context.Context, data *Secret) error
	Update(ctx context.Context, data *Secret) error
	Delete(ctx context.Context, data *Secret) error
}
