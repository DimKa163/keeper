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
	getStoredDataQUERY = `SELECT 
    				id, 
    				created_at,
    				modified_at,
    				user_id, 
    				big_data, 
    				data_type, 
    				payload,
    				dek,
    				path,
    				version,
    				deleted
					FROM data 
					WHERE id = $1 FOR UPDATE`
	getAllStoredDataQUERY = `SELECT 
    				id, 
    				created_at,
    				modified_at,
    				user_id, 
    				big_data, 
    				data_type, 
    				payload,
    				dek, 
    				path,
    				version,
    				deleted
					FROM data 
					WHERE user_id = $1 AND version > $2
					ORDER BY modified_at ASC`
	insertStoredDataQUERY = `INSERT INTO data (
				 	id,
                  	modified_at,
    				user_id, 
    				big_data, 
    				data_type, 
    				payload, 
    				dek, 
                  	path,
    				version,
                  	deleted)
    				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	updateStoredDataQUERY = `UPDATE data 
							SET
							user_id = $2,
							big_data = $3,
							data_type = $4,
							payload = $5,
							dek = $6,
							version = $7,
							modified_at = $8
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
	var userID guid.Guid
	var typeStr string
	var bigData bool
	var payload []byte
	var dek []byte
	var path string
	var version int32
	var deleted bool
	if err := sdr.db.QueryRow(ctx, getStoredDataQUERY, dataID).
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

func (sdr *DataRepository) GetAll(ctx context.Context, userID guid.Guid, greater int32) ([]*domain.Data, error) {
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

func (sdr *DataRepository) Insert(ctx context.Context, data *domain.Data) error {
	if _, err := sdr.db.Exec(
		ctx,
		insertStoredDataQUERY,
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

func (sdr *DataRepository) Update(ctx context.Context, data *domain.Data) error {
	if _, err := sdr.db.Exec(
		ctx,
		updateStoredDataQUERY,
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
