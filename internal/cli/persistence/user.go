package persistence

import (
	"context"
	"database/sql"

	"github.com/DimKa163/keeper/internal/cli/core"
)

const (
	getUserByLoginStmt = `SELECT id, username, password FROM users WHERE username = ?`
	insertUserStmt     = `INSERT INTO users (id, username, password, salt) VALUES (?, ?, ?, ?)`
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (rep *UserRepository) Get(ctx context.Context, login string) (*core.User, error) {
	var user core.User
	if err := rep.db.QueryRowContext(ctx, getUserByLoginStmt, login).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Salt,
	); err != nil {
		return nil, err
	}
	return &user, nil
}

func (rep *UserRepository) Insert(ctx context.Context, user *core.User) error {
	if _, err := rep.db.ExecContext(ctx, insertUserStmt, user.ID, user.Username, user.Password, user.Salt); err != nil {
		return err
	}
	return nil
}
