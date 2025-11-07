package persistence

import (
	"context"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/shared/db"
)

const (
	getUserByLoginQUERY = "SELECT id, created_at, login, password, salt, encrypt_salt FROM users WHERE login = $1"
	existQUERY          = "SELECT EXISTS(SELECT id FROM users WHERE login = $1)"
	insertQueryQUERY    = "INSERT INTO users (created_at, login, password, salt, encrypt_salt) VALUES ($1, $2, $3, $4, $5) RETURNING id"
)

type userRepository struct {
	db db.QueryExecutor
}

func NewUserRepository(db db.QueryExecutor) *userRepository {
	return &userRepository{db: db}
}

func (ur *userRepository) Get(ctx context.Context, login string) (*domain.User, error) {
	var user domain.User
	if err := ur.db.QueryRow(ctx, getUserByLoginQUERY, login).Scan(&user.ID,
		&user.CreatedAt,
		&user.Login,
		&user.Password,
		&user.Salt,
		&user.EncryptSalt); err != nil {
		return nil, err
	}
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
		user.CreatedAt,
		user.Login,
		user.Password,
		user.Salt,
		user.EncryptSalt,
	); err != nil {
		return err
	}
	return nil
}
