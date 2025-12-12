// Package persistence db tools
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/DimKa163/keeper/internal/cli/core"
)

const (
	conflictExistsStmt       = `SELECT EXISTS(SELECT id FROM conflicts)`
	getAllNotSolvedConflicts = `SELECT id,
					created_at,
					modified_at,
					record_id,
					local,
    				remote
					FROM conflicts`
	insertConflict = `INSERT INTO conflicts (record_id, local, remote) VALUES (?, ?, ?)`

	conflictCount = `SELECT COUNT(*) FROM conflicts`

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

func GetAllConflict(ctx context.Context, db *sql.DB) ([]*core.Conflict, error) {
	row, err := db.QueryContext(ctx, getAllNotSolvedConflicts)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	conflicts := make([]*core.Conflict, 0)
	for row.Next() {
		var id int32
		var createdAt, modifiedAt sql.NullTime
		var recordID string
		var local, remote []byte
		if err = row.Scan(&id, &createdAt, &modifiedAt, &recordID, &local, &remote); err != nil {
			return nil, err
		}
		conflict := &core.Conflict{
			ID:       id,
			RecordID: recordID,
		}

		conflicts = append(conflicts, conflict)
		if createdAt.Valid {
			conflict.CreatedAt = createdAt.Time
		}
		if modifiedAt.Valid {
			conflict.ModifiedAt = modifiedAt.Time
		}
		var localItem core.ConflictItem
		if err = json.Unmarshal(local, &localItem); err != nil {
			return nil, err
		}
		conflict.Local = &localItem
		var remoteItem core.ConflictItem
		if err = json.Unmarshal(remote, &remoteItem); err != nil {
			return nil, err
		}
		conflict.Remote = &remoteItem
	}
	return conflicts, nil
}

func DeleteConflict(ctx context.Context, db *sql.DB, id int32) error {
	if _, err := db.ExecContext(ctx, deleteConflictStmt, id); err != nil {
		return err
	}
	return nil
}

func TxGetConflictCount(ctx context.Context, db *sql.Tx) (int64, error) {
	var count int64
	if err := db.QueryRowContext(ctx, conflictCount).Scan(&count); err != nil {
		return -1, err
	}
	return count, nil
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
