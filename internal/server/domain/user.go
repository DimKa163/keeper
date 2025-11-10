package domain

import (
	"context"
	"time"

	"github.com/beevik/guid"
)

// User data owner
type User struct {
	ID        guid.Guid
	CreatedAt *time.Time
	Login     string
	Password  []byte
	Salt      []byte
}

func NewUser(login string, pass, salt []byte) *User {
	return &User{
		Login:    login,
		Password: pass,
		Salt:     salt,
	}
}

type UserRepository interface {
	Get(ctx context.Context, login string) (*User, error)
	Exist(ctx context.Context, login string) (bool, error)
	Insert(ctx context.Context, user *User) error
}

type UserService interface {
	// Login authenticate users
	Login(ctx context.Context, login string, password string) (string, error)
	// Register create new user
	Register(ctx context.Context, login string, password string) (string, error)
}
