package persistence

import (
	"context"
	"database/sql"
	"github.com/DimKa163/keeper/internal/cli/core"
)

const (
	getStateByNameStmt = `SELECT id, value FROM sync_state WHERE id = ?`
	upsertStateStmt    = `INSERT INTO sync_state(id, value) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET value=excluded.value`
)

func GetState(ctx context.Context, db *sql.DB, name string) (*core.SyncState, error) {
	var syncState core.SyncState
	if err := db.QueryRowContext(ctx, getStateByNameStmt, name).Scan(&syncState.ID, &syncState.Value); err != nil {
		return nil, err
	}
	return &syncState, nil
}

func TxGetState(ctx context.Context, db *sql.Tx, name string) (*core.SyncState, error) {
	var syncState core.SyncState
	if err := db.QueryRowContext(ctx, getStateByNameStmt, name).Scan(&syncState.ID, &syncState.Value); err != nil {
		return nil, err
	}
	return &syncState, nil
}

func SaveState(ctx context.Context, db *sql.Tx, state *core.SyncState) error {
	if _, err := db.ExecContext(ctx, upsertStateStmt, state.ID, state.Value); err != nil {
		return err
	}
	return nil
}
