package app

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"os"

	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	shared "github.com/DimKa163/keeper/internal/datatool"
	"github.com/beevik/guid"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (us *UserService) Register(ctx context.Context, key string) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	salt, err := shared.GenerateSalt()
	if err != nil {
		return err
	}
	hash := shared.Hash([]byte(key), salt, 2, 64, 32, 2)
	user := &core.User{
		ID:       guid.New().String(),
		Username: hostname,
		Password: hash,
		Salt:     salt,
	}
	if err = persistence.InsertUser(ctx, us.db, user); err != nil {
		return err
	}
	return nil
}

func (us *UserService) Auth(ctx context.Context, pass string) ([]byte, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	user, err := persistence.GetUser(ctx, us.db, hostname)
	if err != nil {
		return nil, err
	}
	hash := shared.Hash([]byte(pass), user.Salt, 2, 64, 32, 2)
	if !shared.Compare(hash, user.Password) {
		return nil, errors.New("invalid password")
	}
	key := sha256.Sum256([]byte(pass))
	return key[:], nil
}
