package persistence

import (
	"context"
	"database/sql"
	"github.com/DimKa163/keeper/internal/cli/core"
)

const (
	conflictExistsStmt       = `SELECT EXISTS(SELECT id FROM conflicts)`
	getAllNotSolvedConflicts = `SELECT id,
					created_at,
					modified_at,
					record_id,
					local,
    				remote,
    				solved
					FROM conflicts`
	insertConflict = `INSERT INTO conflicts (record_id, local, remote) VALUES (?, ?, ?)`

	deleteConflictStmt = `DELETE FROM conflicts WHERE id=?`
)

func TxConflictExist(ctx context.Context, db *sql.Tx) (bool, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, conflictExistsStmt).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
func ConflictExist(ctx context.Context, db *sql.DB) (bool, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, conflictExistsStmt).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func TxInsertConflict(ctx context.Context, tx *sql.Tx, conflict *core.Conflict) error {
	local, err := conflict.MarshalLocal()
	if err != nil {
		return err
	}
	remote, err := conflict.MarshalRemote()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, insertConflict, conflict.RecordID, local, remote)
	if err != nil {
		return err
	}
	return nil
}
