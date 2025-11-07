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

func (us *UserService) Login(ctx context.Context, login, password string) (string, []byte, error) {
	repository := us.unitOfWork.UserRepository()
	exist, err := repository.Exist(ctx, login)
	if err != nil {
		return "", nil, err
	}
	if !exist {
		return "", nil, ErrUserNotFound
	}
	user, err := repository.Get(ctx, login)
	if err != nil {
		return "", nil, err
	}
	if err := us.authService.Authenticate([]byte(password), user.Password, user.Salt); err != nil {
		return "", nil, err
	}
	token, err := us.authEngine.GenerateToken(user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, user.EncryptSalt, nil
}

func (us *UserService) Register(ctx context.Context, login, password string) (string, []byte, error) {
	repository := us.unitOfWork.UserRepository()
	exist, err := repository.Exist(ctx, login)
	if err != nil {
		return "", nil, err
	}
	if exist {
		return "", nil, ErrLoginAlreadyExists
	}
	pwd, salt, err := us.authService.GenerateHash([]byte(password))
	if err != nil {
		return "", nil, err
	}
	//TODO пересмотреть
	encSalt, err := us.authService.GenerateSalt()
	if err != nil {
		return "", nil, err
	}
	user := domain.NewUser(login, pwd, salt, encSalt)
	if err := repository.Insert(ctx, user); err != nil {
		return "", nil, err
	}
	user, err = repository.Get(ctx, login)
	if err != nil {
		return "", nil, err
	}
	token, err := us.authEngine.GenerateToken(user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, user.EncryptSalt, nil
}
