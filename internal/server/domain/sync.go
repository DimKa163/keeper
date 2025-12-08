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
	Insert(ctx context.Context, state *SyncState) error
	Update(ctx context.Context, syncState *SyncState) error
}
