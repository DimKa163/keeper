package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/shared/db"
	"github.com/beevik/guid"
	"github.com/jackc/pgx/v5"
)

const (
	getStoredDataQUERY = `SELECT 
    				id, 
    				created_at,
    				modified_at,
    				name, 
    				user_id, 
    				large, 
    				data_type, 
    				payload, 
    				payload_nonce, 
    				dek, 
    				dek_nonce, 
    				version,
    				deleted
					FROM data 
					WHERE id = $1 FOR UPDATE`
	getAllStoredDataQUERY = `SELECT 
    				id, 
    				created_at,
    				modified_at,
    				name, 
    				user_id, 
    				large, 
    				data_type, 
    				payload, 
    				payload_nonce, 
    				dek, 
    				dek_nonce, 
    				version,
    				deleted
					FROM data 
					WHERE user_id = $1 AND modified_at > $2
					ORDER BY modified_at ASC`
	insertStoredDataQUERY = `INSERT INTO data (
				 	id,
    				name, 
    				user_id, 
    				large, 
    				data_type, 
    				payload, 
    				payload_nonce, 
    				dek, 
    				dek_nonce, 
    				version)
    				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	updateStoredDataQUERY = `UPDATE data 
							SET
							name = $2,
							user_id = $3,
							large = $4,
							payload_type = $5,
							payload = $6,
							data_nonce = $7,
							dek = $8,
							dek_nonce = $9,
							version = $10
							modified_at = $11
							WHERE id = $1`
	deleteStoredDataQUERY = `UPDATE data SET deleted = $2, version=$3 WHERE id = $1`
)

type DataRepository struct {
	db db.QueryExecutor
}

func NewStoredDataRepository(db db.QueryExecutor) *DataRepository {
	return &DataRepository{db: db}
}

func (sdr *DataRepository) Get(ctx context.Context, dataID guid.Guid) (*domain.Data, error) {
	var storedData domain.Data
	var id guid.Guid
	var createdAt sql.NullTime
	var modifiedAt sql.NullTime
	var name string
	var userID guid.Guid
	var typeStr string
	var large bool
	var payload []byte
	var payloadNonce []byte
	var dek []byte
	var dekNonce []byte
	var version int32
	var deleted bool
	if err := sdr.db.QueryRow(ctx, getStoredDataQUERY, dataID).
		Scan(&id,
			&createdAt,
			&modifiedAt,
			&name,
			&userID,
			&large,
			&typeStr,
			&payload,
			&payloadNonce,
			&dek,
			&dekNonce,
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
	storedData.Name = name
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
	storedData.Large = large
	storedData.Payload = payload
	storedData.PayloadNonce = payloadNonce
	storedData.Dek = dek
	storedData.DekNonce = dekNonce
	storedData.Version = version
	storedData.Deleted = deleted
	return &storedData, nil
}

func (sdr *DataRepository) GetAll(ctx context.Context, userID guid.Guid, greater time.Time) ([]*domain.Data, error) {
	row, err := sdr.db.Query(ctx, getAllStoredDataQUERY, userID, greater)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	slice := make([]*domain.Data, 0)
	for row.Next() {
		var data domain.Data
		var id guid.Guid
		var createdAt sql.NullTime
		var modifiedAt sql.NullTime
		var name string
		var usID guid.Guid
		var typeStr string
		var large bool
		var payload []byte
		var payloadNonce []byte
		var dek []byte
		var dekNonce []byte
		var version int32
		var deleted bool
		if err := row.Scan(&id,
			&createdAt,
			&modifiedAt,
			&name,
			&usID,
			&large,
			&typeStr,
			&payload,
			&payloadNonce,
			&dek,
			&dekNonce,
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
		data.Name = name
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
		data.Large = large
		data.Payload = payload
		data.PayloadNonce = payloadNonce
		data.Dek = dek
		data.DekNonce = dekNonce
		data.Version = version
		data.Deleted = deleted
	}
	return slice, nil
}

func (sdr *DataRepository) Insert(ctx context.Context, data *domain.Data) error {
	if _, err := sdr.db.Exec(
		ctx,
		insertStoredDataQUERY,
		data.ID,
		data.Name,
		data.UserID,
		data.Large,
		data.Type,
		data.Payload,
		data.PayloadNonce,
		data.Dek,
		data.DekNonce,
		data.Version,
	); err != nil {
		return err
	}
	return nil
}

func (sdr *DataRepository) Update(ctx context.Context, data *domain.Data) error {
	if _, err := sdr.db.Exec(
		ctx,
		updateStoredDataQUERY,
		data.ID,
		data.Name,
		data.UserID,
		data.Large,
		data.Type,
		data.Payload,
		data.PayloadNonce,
		data.Dek,
		data.DekNonce,
		data.Version,
		data.ModifiedAt,
	); err != nil {
		return err
	}
	return nil
}

func (sdr *DataRepository) Delete(ctx context.Context, data *domain.Data) error {
	if _, err := sdr.db.Exec(
		ctx,
		deleteStoredDataQUERY,
		data.ID,
		data.Deleted,
		data.Version,
	); err != nil {
		return err
	}
	return nil
}
