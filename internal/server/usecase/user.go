package usecase

import (
	"context"
	"errors"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/domain/auth"
)

var (
	ErrLoginAlreadyExists = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type UserService struct {
	unitOfWork  domain.UnitOfWork
	authService auth.AuthService
	authEngine  auth.Engine
}

func NewUserService(unitOfWork domain.UnitOfWork, authService auth.AuthService, engine auth.Engine) *UserService {
	return &UserService{unitOfWork: unitOfWork, authService: authService, authEngine: engine}
}

func (us *UserService) Login(ctx context.Context, login, password string) (string, error) {
	repository := us.unitOfWork.UserRepository()
	exist, err := repository.Exist(ctx, login)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", ErrUserNotFound
	}
	user, err := repository.Get(ctx, login)
	if err != nil {
		return "", err
	}
	if err := us.authService.Authenticate([]byte(password), user.Password, user.Salt); err != nil {
		return "", err
	}
	token, err := us.authEngine.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (us *UserService) Register(ctx context.Context, login, password string) (string, error) {
	if err := us.unitOfWork.Tx(ctx, func(ctx context.Context, work domain.UnitOfWork) error {
		repository := work.UserRepository()
		exist, err := repository.Exist(ctx, login)
		if err != nil {
			return err
		}
		if exist {
			return ErrLoginAlreadyExists
		}
		pwd, salt, err := us.authService.GenerateHash([]byte(password))
		if err != nil {
			return err
		}
		user := domain.NewUser(login, pwd, salt)
		if err := repository.Insert(ctx, user); err != nil {
			return err
		}
		user, err = repository.Get(ctx, login)
		if err != nil {
			return err
		}
		var state domain.SyncState
		state.ID = syncTypeName
		state.UserID = user.ID
		state.Value = 0
		if err = work.SyncStateRepository().Insert(ctx, &state); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return "", err
	}
	repository := us.unitOfWork.UserRepository()
	user, err := repository.Get(ctx, login)
	if err != nil {
		return "", err
	}
	token, err := us.authEngine.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}
