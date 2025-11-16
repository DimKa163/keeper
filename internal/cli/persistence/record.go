package persistence

import (
	"context"
	"database/sql"

	"github.com/DimKa163/keeper/internal/cli/core"
)

const (
	recordExistsStmt = `SELECT EXISTS(SELECT id FROM records WHERE id = $1)`
	getAllStmt       = `SELECT id, created_at, modified_at, type, data, data_nonce, dek, dek_nonce, version FROM records
				ORDER BY id
				LIMIT ? OFFSET ?`
	getRecordByIDStmt = `SELECT id, created_at, modified_at, type, data, data_nonce, dek, dek_nonce, version FROM records
			WHERE id = ?`
	insertStmt = `INSERT INTO records (id, created_at, modified_at, type, data, data_nonce, dek, dek_nonce, version) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	updateStmt = `UPDATE records SET data = ?, data_nonce = ?, dek = ?, dek_nonce = ?, version = ?, modified_at = ? WHERE id = ?`

	updateVersionStmt = `UPDATE records SET version = ? WHERE id = ?`

	deleteStmt = `DELETE FROM records WHERE id = ?`
)

type RecordRepository struct {
	db *sql.DB
}

func NewRecordRepository(db *sql.DB) *RecordRepository {
	return &RecordRepository{db: db}
}

func GetAllRecord(ctx context.Context, db *sql.DB, limit, offset int) ([]*core.Record, error) {
	rows, err := db.QueryContext(ctx, getAllStmt, limit, offset)
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
			&r.Data,
			&r.DataNonce,
			&r.Dek,
			&r.DekNonce,
			&r.Version); err != nil {
			return nil, err
		}
		records = append(records, &r)
	}
	return records, nil
}

func TxGetAllRecord(ctx context.Context, tx *sql.Tx, limit, offset int) ([]*core.Record, error) {
	rows, err := tx.QueryContext(ctx, getAllStmt, limit, offset)
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
			&r.Data,
			&r.DataNonce,
			&r.Dek,
			&r.DekNonce,
			&r.Version); err != nil {
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
		&r.Data,
		&r.DataNonce,
		&r.Dek,
		&r.DekNonce,
		&r.Version); err != nil {
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
		&r.Data,
		&r.DataNonce,
		&r.Dek,
		&r.DekNonce,
		&r.Version); err != nil {
		return nil, err
	}
	return &r, nil
}

func InsertRecord(ctx context.Context, db *sql.DB, record *core.Record) error {
	if _, err := db.ExecContext(
		ctx,
		insertStmt,
		record.ID,
		record.CreatedAt,
		record.ModifiedAt,
		record.Type,
		record.Data,
		record.DataNonce,
		record.Dek,
		record.DekNonce,
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
		record.Data,
		record.DataNonce,
		record.Dek,
		record.DekNonce,
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
		record.Data,
		record.DataNonce,
		record.Dek,
		record.DekNonce,
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
		record.Data,
		record.DataNonce,
		record.Dek,
		record.DekNonce,
		record.Version,
		record.ModifiedAt,
		record.ID,
	); err != nil {
		return err
	}
	return nil
}

func (rep *RecordRepository) UpdateVersion(ctx context.Context, id string, version int32) error {
	if _, err := rep.db.ExecContext(
		ctx,
		updateVersionStmt,
		version,
		id,
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
