package domain

import (
	"context"
	"github.com/beevik/guid"
)

type SyncState struct {
	ID     string
	UserID guid.Guid
	Value  int32
}

type SyncStateRepository interface {
	Get(ctx context.Context, id string, user guid.Guid) (*SyncState, error)
	Save(ctx context.Context, syncState *SyncState) error
}
