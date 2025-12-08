package persistence

import (
	"context"
	"database/sql"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/shared/db"
	"github.com/beevik/guid"
)

const (
	getUserByLoginQUERY = "SELECT id, created_at, login, password, salt FROM users WHERE login = $1"
	existQUERY          = "SELECT EXISTS(SELECT id FROM users WHERE login = $1)"
	insertQueryQUERY    = "INSERT INTO users (login, password, salt) VALUES ($1, $2, $3) RETURNING id"
)

type userRepository struct {
	db db.QueryExecutor
}

func NewUserRepository(db db.QueryExecutor) *userRepository {
	return &userRepository{db: db}
}

func (ur *userRepository) Get(ctx context.Context, login string) (*domain.User, error) {
	var user domain.User
	var id guid.Guid
	var createdAt sql.NullTime
	var lgn string
	var pwd []byte
	var salt []byte
	if err := ur.db.QueryRow(ctx, getUserByLoginQUERY, login).Scan(&id,
		&createdAt,
		&lgn,
		&pwd,
		&salt); err != nil {
		return nil, err
	}
	user.ID = id
	if createdAt.Valid {
		user.CreatedAt = &createdAt.Time
	}
	user.Login = login
	user.Password = pwd
	user.Salt = salt
	return &user, nil
}

func (ur *userRepository) Exist(ctx context.Context, login string) (bool, error) {
	var exst bool
	if err := ur.db.QueryRow(ctx, existQUERY, login).Scan(&exst); err != nil {
		return false, err
	}
	return exst, nil
}

func (ur *userRepository) Insert(ctx context.Context, user *domain.User) error {
	if _, err := ur.db.Exec(
		ctx,
		insertQueryQUERY,
		user.Login,
		user.Password,
		user.Salt,
	); err != nil {
		return err
	}
	return nil
}
