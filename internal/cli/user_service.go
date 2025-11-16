package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	shared "github.com/DimKa163/keeper/internal/shared"
	"github.com/beevik/guid"
)

type UserService struct {
	userRepository *persistence.UserRepository
}

func NewUserService(userRepository *persistence.UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (us *UserService) Init(ctx context.Context, console *Console) ([]byte, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	user, err := us.userRepository.Get(ctx, hostname)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	fmt.Print("Master Key: ")

	if errors.Is(err, sql.ErrNoRows) {
		masterKey, err := console.Read()
		if err != nil {
			return nil, err
		}
		salt, err := shared.GenerateSalt()
		if err != nil {
			return nil, err
		}
		//TODO in config
		hash := shared.Hash([]byte(masterKey), salt, 2, 64, 32, 2)
		user = &core.User{
			ID:       guid.New().String(),
			Username: hostname,
			Password: hash,
			Salt:     salt,
		}
		if err = us.userRepository.Insert(ctx, user); err != nil {
			return nil, err
		}
		return hash, nil

	} else {
		masterKey, err := console.Read()
		if err != nil {
			return nil, err
		}
		hash := shared.Hash([]byte(masterKey), user.Salt, 2, 64, 32, 2)
		if !shared.Compare(hash, user.Password) {
			return nil, errors.New("invalid password")
		}
		fmt.Println("Hi, 👋")
		return user.Password, nil
	}
}
