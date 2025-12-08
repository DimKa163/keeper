package persistence

import (
	"context"
	"database/sql"

	"github.com/DimKa163/keeper/internal/cli/core"
)

const (
	getUserByLoginStmt = `SELECT id, username, password, salt FROM users WHERE username = ?`
	insertUserStmt     = `INSERT INTO users (id, username, password, salt) VALUES (?, ?, ?, ?)`
)

func GetUser(ctx context.Context, db *sql.DB, login string) (*core.User, error) {
	var user core.User
	if err := db.QueryRowContext(ctx, getUserByLoginStmt, login).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Salt,
	); err != nil {
		return nil, err
	}
	return &user, nil
}

func InsertUser(ctx context.Context, db *sql.DB, user *core.User) error {
	if _, err := db.ExecContext(ctx, insertUserStmt, user.ID, user.Username, user.Password, user.Salt); err != nil {
		return err
	}
	return nil
}
