package persistence

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/shared/db"
	"github.com/beevik/guid"
	"github.com/jackc/pgx/v5"
)

const (
	getSecretQUERY = `SELECT 
    				id, 
    				created_at,
    				modified_at,
    				user_id, 
    				big_data, 
    				secret_type, 
    				payload,
    				dek,
    				path,
    				version,
    				deleted
					FROM secret 
					WHERE id = $1 FOR UPDATE`
	getAllSecretQUERY = `SELECT 
    				id, 
    				created_at,
    				modified_at,
    				user_id, 
    				big_data, 
    				secret_type, 
    				payload,
    				dek, 
    				path,
    				version,
    				deleted
					FROM secret 
					WHERE user_id = $1 AND version > $2
					ORDER BY modified_at ASC`
	insertSecretQUERY = `INSERT INTO secret (
				 	id,
                  	modified_at,
    				user_id, 
    				big_data, 
    				secret_type, 
    				payload, 
    				dek, 
                  	path,
    				version,
                  	deleted)
    				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	updateSecretQUERY = `UPDATE secret 
							SET
							user_id = $2,
							big_data = $3,
							secret_type = $4,
							payload = $5,
							dek = $6,
							version = $7,
							modified_at = $8
							WHERE id = $1`
	deleteSecretQUERY = `UPDATE secret SET deleted = $2, version=$3 WHERE id = $1`
)

type SecretRepository struct {
	db db.QueryExecutor
}

func NewSecretRepository(db db.QueryExecutor) *SecretRepository {
	return &SecretRepository{db: db}
}

func (sdr *SecretRepository) Get(ctx context.Context, dataID guid.Guid) (*domain.Secret, error) {
	var storedData domain.Secret
	var id guid.Guid
	var createdAt sql.NullTime
	var modifiedAt sql.NullTime
	var userID guid.Guid
	var typeStr string
	var bigData bool
	var payload []byte
	var dek []byte
	var path string
	var version int32
	var deleted bool
	if err := sdr.db.QueryRow(ctx, getSecretQUERY, dataID).
		Scan(&id,
			&createdAt,
			&modifiedAt,
			&userID,
			&bigData,
			&typeStr,
			&payload,
			&dek,
			&path,
			&version,
			&deleted); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}
	storedData.ID = id
	if createdAt.Valid {
		storedData.CreatedAt = createdAt.Time
	}
	if modifiedAt.Valid {
		storedData.ModifiedAt = modifiedAt.Time
	}
	storedData.UserID = userID
	switch typeStr {
	case "login_pass":
		storedData.Type = domain.LoginPassType
	case "text":
		storedData.Type = domain.TextType
	case "bank_card":
		storedData.Type = domain.BankCardType
	case "other":
		storedData.Type = domain.OtherType
	}
	storedData.BigData = bigData
	storedData.Payload = payload
	storedData.Dek = dek
	storedData.Path = path
	storedData.Version = version
	storedData.Deleted = deleted
	return &storedData, nil
}

func (sdr *SecretRepository) GetAll(ctx context.Context, userID guid.Guid, greater int32) ([]*domain.Secret, error) {
	row, err := sdr.db.Query(ctx, getAllSecretQUERY, userID, greater)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	slice := make([]*domain.Secret, 0)
	for row.Next() {
		var data domain.Secret
		var id guid.Guid
		var createdAt sql.NullTime
		var modifiedAt sql.NullTime
		var usID guid.Guid
		var typeStr string
		var bigData bool
		var payload []byte
		var dek []byte
		var path string
		var version int32
		var deleted bool
		if err := row.Scan(&id,
			&createdAt,
			&modifiedAt,
			&usID,
			&bigData,
			&typeStr,
			&payload,
			&dek,
			&path,
			&version,
			&deleted); err != nil {
			return nil, err
		}
		slice = append(slice, &data)
		data.ID = id
		if createdAt.Valid {
			data.CreatedAt = createdAt.Time
		}
		if modifiedAt.Valid {
			data.ModifiedAt = modifiedAt.Time
		}
		data.UserID = userID
		switch typeStr {
		case "login_pass":
			data.Type = domain.LoginPassType
		case "text":
			data.Type = domain.TextType
		case "bank_card":
			data.Type = domain.BankCardType
		case "other":
			data.Type = domain.OtherType
		}
		data.BigData = bigData
		data.Payload = payload
		data.Dek = dek
		data.Path = path
		data.Version = version
		data.Deleted = deleted
	}
	return slice, nil
}

func (sdr *SecretRepository) Insert(ctx context.Context, data *domain.Secret) error {
	if _, err := sdr.db.Exec(
		ctx,
		insertSecretQUERY,
		data.ID,
		data.ModifiedAt,
		data.UserID,
		data.BigData,
		data.Type,
		data.Payload,
		data.Dek,
		data.Path,
		data.Version,
		data.Deleted,
	); err != nil {
		return err
	}
	return nil
}

func (sdr *SecretRepository) Update(ctx context.Context, data *domain.Secret) error {
	if _, err := sdr.db.Exec(
		ctx,
		updateSecretQUERY,
		data.ID,
		data.UserID,
		data.BigData,
		data.Type,
		data.Payload,
		data.Dek,
		data.Version,
		data.ModifiedAt,
	); err != nil {
		return err
	}
	return nil
}

func (sdr *SecretRepository) Delete(ctx context.Context, data *domain.Secret) error {
	if _, err := sdr.db.Exec(
		ctx,
		deleteSecretQUERY,
		data.ID,
		data.Deleted,
		data.Version,
	); err != nil {
		return err
	}
	return nil
}
