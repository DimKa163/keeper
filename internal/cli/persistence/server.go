package persistence

import (
	"context"
	"database/sql"

	"github.com/DimKa163/keeper/internal/cli/core"
)

const (
	insertServerStmt = `INSERT INTO servers (address, login, password, active) VALUES (?, ?, ?, ?)`
	getServerStmt    = `SELECT id, address, login, password, active FROM servers WHERE active = $1 LIMIT 1`

	updateActiveServerStmt = `UPDATE servers SET active = $1 WHERE id = $2`
)

func InsertServer(ctx context.Context, db *sql.DB, address, login, pass string, active bool) error {
	if _, err := db.ExecContext(ctx, insertServerStmt, address, login, pass, active); err != nil {
		return err
	}
	return nil
}

func GetServer(ctx context.Context, db *sql.DB, active bool) (*core.Server, error) {
	var server core.Server
	if err := db.QueryRowContext(ctx, getServerStmt, active).Scan(
		&server.ID,
		&server.Address,
		&server.Login,
		&server.Password,
		&server.Active,
	); err != nil {
		return nil, err
	}
	return &server, nil
}

func UpdateServer(ctx context.Context, db *sql.DB, id int32, active bool) error {
	if _, err := db.Exec(updateActiveServerStmt, active, id); err != nil {
		return err
	}
	return nil
}
