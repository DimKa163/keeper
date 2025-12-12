package persistence

import (
	"context"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/shared/db"
	"github.com/beevik/guid"
)

const (
	getStateQUERY    = `SELECT id, user_id, value FROM sync_state WHERE id = $1 AND user_id = $2 FOR UPDATE;`
	insertStateQuery = `INSERT INTO sync_state VALUES ($1, $2, $3);`
	updateStateQUERY = `UPDATE sync_state SET value = $1 WHERE id = $2 and user_id = $3;`
)

type SyncStateRepository struct {
	db db.QueryExecutor
}

func NewSyncStateRepository(db db.QueryExecutor) *SyncStateRepository {
	return &SyncStateRepository{db: db}
}

func (sr *SyncStateRepository) Get(ctx context.Context, id string, userID guid.Guid) (*domain.SyncState, error) {
	var syncState domain.SyncState
	if err := sr.db.QueryRow(ctx, getStateQUERY, id, userID).Scan(&syncState.ID, &syncState.UserID, &syncState.Value); err != nil {
		return nil, err
	}
	return &syncState, nil
}

func (sr *SyncStateRepository) Insert(ctx context.Context, syncState *domain.SyncState) error {
	if _, err := sr.db.Exec(ctx, insertStateQuery, syncState.ID, syncState.UserID, syncState.Value); err != nil {
		return err
	}
	return nil
}

func (sr *SyncStateRepository) Update(ctx context.Context, syncState *domain.SyncState) error {
	if _, err := sr.db.Exec(ctx, updateStateQUERY, syncState.Value, syncState.ID, syncState.UserID); err != nil {
		return err
	}
	return nil
}
