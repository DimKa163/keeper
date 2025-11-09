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
    				name, 
    				user_id, 
    				large, 
    				data_type, 
    				data, 
    				data_nonce, 
    				dek, 
    				dek_nonce, 
    				version 
					FROM stored_data 
					WHERE id = $1 FOR UPDATE`
	getAllStoredDataQUERY = `SELECT 
    				id, 
    				created_at, 
    				name, 
    				user_id, 
    				large, 
    				data_type, 
    				data, 
    				data_nonce, 
    				dek, 
    				dek_nonce, 
    				version 
					FROM stored_data 
					WHERE user_id = $1
					ORDER BY id
					LIMIT $2
					OFFSET $3`
	insertStoredDataQUERY = `INSERT INTO stored_data (
				 	id,
    				name, 
    				user_id, 
    				large, 
    				data_type, 
    				data, 
    				data_nonce, 
    				dek, 
    				dek_nonce, 
    				version)
    				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	updateStoredDataQUERY = `UPDATE stored_data 
							SET
							name = $2,
							user_id = $3,
							large = $4,
							data_type = $5,
							data = $6,
							data_nonce = $7,
							dek = $8,
							dek_nonce = $9,
							version = $10
							WHERE id = $1`
	deleteStoredDataQUERY = `DELETE FROM stored_data WHERE id = $1`
)

type storedDataRepository struct {
	db db.QueryExecutor
}

func NewStoredDataRepository(db db.QueryExecutor) *storedDataRepository {
	return &storedDataRepository{db: db}
}

func (sdr *storedDataRepository) Get(ctx context.Context, dataID guid.Guid) (*domain.StoredData, error) {
	var storedData domain.StoredData
	var id guid.Guid
	var createdAt sql.NullTime
	var name string
	var userID guid.Guid
	var typeStr string
	var large bool
	var data []byte
	var dataNonce []byte
	var dek []byte
	var dekNonce []byte
	var version int32
	if err := sdr.db.QueryRow(ctx, getStoredDataQUERY, dataID).
		Scan(&id,
			&createdAt,
			&name,
			&userID,
			&large,
			&typeStr,
			&data,
			&dataNonce,
			&dek,
			&dekNonce,
			&version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}
	storedData.ID = id
	if createdAt.Valid {
		storedData.CreatedAt = createdAt.Time
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
	storedData.Data = data
	storedData.DataNonce = dataNonce
	storedData.Dek = dek
	storedData.DekNonce = dekNonce
	storedData.Version = version
	return &storedData, nil
}

func (sdr *storedDataRepository) GetAll(ctx context.Context, userID guid.Guid, limit, skip int) ([]*domain.StoredData, error) {
	row, err := sdr.db.Query(ctx, getAllStoredDataQUERY, userID, limit, skip)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	index := 0
	slice := make([]*domain.StoredData, limit)
	for row.Next() {
		var storedData domain.StoredData
		var id guid.Guid
		var createdAt sql.NullTime
		var name string
		var usID guid.Guid
		var typeStr string
		var large bool
		var data []byte
		var dataNonce []byte
		var dek []byte
		var dekNonce []byte
		var version int32
		if err := row.Scan(&id,
			&createdAt,
			&name,
			&usID,
			&large,
			&typeStr,
			&data,
			&dataNonce,
			&dek,
			&dekNonce,
			&version); err != nil {
			return nil, err
		}
		slice[index] = &storedData
		storedData.ID = id
		if createdAt.Valid {
			storedData.CreatedAt = createdAt.Time
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
		storedData.Data = data
		storedData.DataNonce = dataNonce
		storedData.Dek = dek
		storedData.DekNonce = dekNonce
		storedData.Version = version
		index++
	}
	return slice, nil
}

func (sdr *storedDataRepository) Insert(ctx context.Context, data *domain.StoredData) error {
	if _, err := sdr.db.Exec(
		ctx,
		insertStoredDataQUERY,
		data.ID,
		data.Name,
		data.UserID,
		data.Large,
		data.Type,
		data.Data,
		data.DataNonce,
		data.Dek,
		data.DekNonce,
		data.Version,
	); err != nil {
		return err
	}
	return nil
}

func (sdr *storedDataRepository) Update(ctx context.Context, data *domain.StoredData) error {
	if _, err := sdr.db.Exec(
		ctx,
		updateStoredDataQUERY,
		data.ID,
		data.Name,
		data.UserID,
		data.Large,
		data.Type,
		data.Data,
		data.DataNonce,
		data.Dek,
		data.DekNonce,
		data.Version,
	); err != nil {
		return err
	}
	return nil
}

func (sdr *storedDataRepository) Delete(ctx context.Context, id guid.Guid) error {
	if _, err := sdr.db.Exec(
		ctx,
		deleteStoredDataQUERY,
		id,
	); err != nil {
		return err
	}
	return nil
}
