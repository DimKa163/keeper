package domain

import (
	"context"
	"time"
)

// User data owner
type User struct {
	ID          int64
	CreatedAt   time.Time
	Login       string
	Password    []byte
	Salt        []byte
	EncryptSalt []byte
}

func NewUser(login string, pass, salt, encSalt []byte) *User {
	return &User{
		CreatedAt:   time.Now(),
		Login:       login,
		Password:    pass,
		Salt:        salt,
		EncryptSalt: encSalt,
	}
}

type UserRepository interface {
	Get(ctx context.Context, login string) (*User, error)
	Exist(ctx context.Context, login string) (bool, error)
	Insert(ctx context.Context, user *User) error
}

type UserService interface {
	// Login authenticate users
	Login(ctx context.Context, login string, password string) (string, []byte, error)
	// Register create new user
	Register(ctx context.Context, login string, password string) (string, []byte, error)
}
