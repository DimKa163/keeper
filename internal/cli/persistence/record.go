package persistence

import (
	"context"
	"database/sql"

	"github.com/DimKa163/keeper/internal/cli/core"
)

const (
	recordExistsStmt = `SELECT EXISTS(SELECT id FROM records WHERE id = $1)`
	getAllStmt       = `SELECT id, created_at, modified_at, type, big_data, data, dek, version, deleted, corrupted FROM records
				WHERE deleted = ? and corrupted = ?
				ORDER BY id
				LIMIT ? OFFSET ?`
	getRecordByIDStmt = `SELECT id, created_at, modified_at, type, big_data, data,  dek, version, deleted, corrupted FROM records
			WHERE id = ?`
	insertStmt = `INSERT INTO records (id, created_at, modified_at, type, big_data, data,  dek,  version) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	updateStmt = `UPDATE records SET big_data = ?, data = ?, dek = ?, version = ?, deleted = ?, corrupted= ?, modified_at = ? WHERE id = ?`

	updateVersionStmt = `UPDATE records SET version = ? WHERE id = ?`

	deleteStmt                 = `DELETE FROM records WHERE id = ?`
	getAllRecordGreaterVersion = `SELECT id, created_at, modified_at, type, big_data, data, dek, deleted, version, corrupted FROM records
	WHERE version > ? AND corrupted = ?`
	updateCorruptedStmt = `UPDATE records SET corrupted = ? WHERE id = ?`
)

func GetAllRecord(ctx context.Context, db *sql.DB, limit, offset int32) ([]*core.Record, error) {
	rows, err := db.QueryContext(ctx, getAllStmt, false, false, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]*core.Record, 0, limit)
	for rows.Next() {
		var r core.Record
		if err = rows.Scan(&r.ID,
			&r.CreatedAt,
			&r.ModifiedAt,
			&r.Type,
			&r.BigData,
			&r.Data,
			&r.Dek,
			&r.Version,
			&r.Deleted,
			&r.Corrupted); err != nil {
			return nil, err
		}
		records = append(records, &r)
	}
	return records, nil
}

func TxGetAllRecord(ctx context.Context, tx *sql.Tx, limit, offset int) ([]*core.Record, error) {
	rows, err := tx.QueryContext(ctx, getAllStmt, false, false, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]*core.Record, 0, limit)
	for rows.Next() {
		var r core.Record
		if err = rows.Scan(&r.ID,
			&r.CreatedAt,
			&r.ModifiedAt,
			&r.Type,
			&r.BigData,
			&r.Data,
			&r.Dek,
			&r.Version,
			&r.Deleted); err != nil {
			return nil, err
		}
		records = append(records, &r)
	}
	return records, nil
}

func TxGetAllRecordGreater(ctx context.Context, tx *sql.Tx, version int32) ([]*core.Record, error) {
	rows, err := tx.QueryContext(ctx, getAllRecordGreaterVersion, version, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]*core.Record, 0)
	for rows.Next() {
		var r core.Record
		if err = rows.Scan(&r.ID,
			&r.CreatedAt,
			&r.ModifiedAt,
			&r.Type,
			&r.BigData,
			&r.Data,
			&r.Dek,
			&r.Deleted,
			&r.Version,
			&r.Corrupted); err != nil {
			return nil, err
		}
		records = append(records, &r)
	}
	return records, nil
}

func TxExists(ctx context.Context, db *sql.Tx, id string) (bool, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, recordExistsStmt, id).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func Exists(ctx context.Context, db *sql.DB, id string) (bool, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, recordExistsStmt, id).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
func GetRecordByID(ctx context.Context, db *sql.DB, id string) (*core.Record, error) {
	var r core.Record
	if err := db.QueryRowContext(ctx, getRecordByIDStmt, id).Scan(&r.ID,
		&r.CreatedAt,
		&r.ModifiedAt,
		&r.Type,
		&r.BigData,
		&r.Data,
		&r.Dek,
		&r.Version,
		&r.Deleted,
		&r.Corrupted); err != nil {
		return nil, err
	}
	return &r, nil
}
func TxGetRecordByID(ctx context.Context, tx *sql.Tx, id string) (*core.Record, error) {
	var r core.Record
	if err := tx.QueryRowContext(ctx, getRecordByIDStmt, id).Scan(&r.ID,
		&r.CreatedAt,
		&r.ModifiedAt,
		&r.Type,
		&r.BigData,
		&r.Data,
		&r.Dek,
		&r.Version,
		&r.Deleted,
		&r.Corrupted); err != nil {
		return nil, err
	}
	return &r, nil
}

func InsertRecord(ctx context.Context, db *sql.DB, record *core.Record) error {
	if _, err := db.ExecContext(
		ctx,
		insertStmt,
		record.ID,
		record.Type,
		record.BigData,
		record.Data,
		record.Dek,
		record.Version,
	); err != nil {
		return err
	}
	return nil
}

func TxInsertRecord(ctx context.Context, db *sql.Tx, record *core.Record) error {
	if _, err := db.ExecContext(
		ctx,
		insertStmt,
		record.ID,
		record.CreatedAt,
		record.ModifiedAt,
		record.Type,
		record.BigData,
		record.Data,
		record.Dek,
		record.Version,
	); err != nil {
		return err
	}
	return nil
}

func UpdateRecord(ctx context.Context, db *sql.DB, record *core.Record) error {
	if _, err := db.ExecContext(
		ctx,
		updateStmt,
		record.BigData,
		record.Data,
		record.Dek,
		record.Version,
		record.Deleted,
		record.Corrupted,
		record.ModifiedAt,
		record.ID,
	); err != nil {
		return err
	}
	return nil
}
func TxUpdateRecord(ctx context.Context, db *sql.Tx, record *core.Record) error {
	if _, err := db.ExecContext(
		ctx,
		updateStmt,
		record.BigData,
		record.Data,
		record.Dek,
		record.Version,
		record.Deleted,
		record.Corrupted,
		record.ModifiedAt,
		record.ID,
	); err != nil {
		return err
	}
	return nil
}

func DeleteRecord(ctx context.Context, db *sql.DB, id string) error {
	if _, err := db.ExecContext(
		ctx,
		deleteStmt,
		id,
	); err != nil {
		return err
	}
	return nil
}

func TxDeleteRecord(ctx context.Context, db *sql.Tx, id string) error {
	if _, err := db.ExecContext(
		ctx,
		deleteStmt,
		id,
	); err != nil {
		return err
	}
	return nil
}
func CorruptRecord(ctx context.Context, db *sql.DB, val bool, id string) error {
	if _, err := db.ExecContext(
		ctx,
		updateCorruptedStmt,
		val,
		id,
	); err != nil {
		return err
	}
	return nil
}

func TxCorruptRecord(ctx context.Context, db *sql.Tx, val bool, id string) error {
	if _, err := db.ExecContext(
		ctx,
		updateCorruptedStmt,
		val,
		id,
	); err != nil {
		return err
	}
	return nil
}
