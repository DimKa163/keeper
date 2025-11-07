package usecase

import (
	"context"
	"testing"

	"github.com/DimKa163/keeper/internal/mocks"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/domain/auth"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUserService_Register_Successfully(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUow := mocks.NewMockUnitOfWork(ctrl)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockAuthEngine := mocks.NewMockEngine(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	userService := createUserService(mockUow, mockAuthService, mockAuthEngine)

	login := "dima"
	password := "qwerty"
	user := &domain.User{
		ID:    1,
		Login: login,

		Password:    []byte(password),
		Salt:        []byte("salt"),
		EncryptSalt: []byte("encrypt_salt"),
	}
	token := "token"
	mockUow.EXPECT().UserRepository().Return(mockRepo).AnyTimes()
	mockRepo.EXPECT().Exist(ctx, login).Return(false, nil)
	mockAuthService.EXPECT().GenerateHash([]byte(password)).Return(user.Password, user.Salt, nil)
	mockAuthService.EXPECT().GenerateSalt().Return(user.EncryptSalt, nil)
	mockRepo.EXPECT().Insert(ctx, &domain.User{Login: login, Password: user.Password, Salt: user.Salt, EncryptSalt: user.EncryptSalt}).Return(nil)
	mockRepo.EXPECT().Get(ctx, login).Return(user, nil)
	mockAuthEngine.EXPECT().GenerateToken(user.ID).Return(token, nil)
	token, encryptSalt, err := userService.Register(ctx, login, password)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, encryptSalt)
	assert.Equal(t, token, token)
}

func TestUserService_Register_FailToCreateUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUow := mocks.NewMockUnitOfWork(ctrl)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockAuthEngine := mocks.NewMockEngine(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	userService := createUserService(mockUow, mockAuthService, mockAuthEngine)

	login := "dima"
	password := "qwerty"
	user := &domain.User{
		ID:    1,
		Login: login,

		Password:    []byte(password),
		Salt:        []byte("salt"),
		EncryptSalt: []byte("encrypt_salt"),
	}
	token := "token"
	mockUow.EXPECT().UserRepository().Return(mockRepo).AnyTimes()
	mockRepo.EXPECT().Exist(ctx, login).Return(true, nil)
	mockAuthService.EXPECT().GenerateHash([]byte(password)).Return(user.Password, user.Salt, nil).Times(0)
	mockAuthService.EXPECT().GenerateSalt().Return(user.EncryptSalt, nil).Times(0)
	mockRepo.EXPECT().Insert(ctx, &domain.User{Login: login, Password: user.Password, Salt: user.Salt, EncryptSalt: user.EncryptSalt}).Return(nil).Times(0)
	mockRepo.EXPECT().Get(ctx, login).Return(user, nil).Times(0)
	mockAuthEngine.EXPECT().GenerateToken(user.ID).Return(token, nil).Times(0)
	token, encryptSalt, err := userService.Register(ctx, login, password)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrLoginAlreadyExists)
	assert.Empty(t, token)
	assert.Nil(t, encryptSalt)
}

func TestUserService_Login_ShouldBeSuccess(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUow := mocks.NewMockUnitOfWork(ctrl)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockAuthEngine := mocks.NewMockEngine(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	userService := createUserService(mockUow, mockAuthService, mockAuthEngine)

	login := "dima"
	password := "qwerty"
	user := &domain.User{
		ID:    1,
		Login: login,

		Password:    []byte(password),
		Salt:        []byte("salt"),
		EncryptSalt: []byte("encrypt_salt"),
	}
	token := "token"
	mockUow.EXPECT().UserRepository().Return(mockRepo).AnyTimes()
	mockRepo.EXPECT().Exist(ctx, login).Return(true, nil)
	mockRepo.EXPECT().Get(ctx, login).Return(user, nil).Times(1)
	mockAuthService.EXPECT().Authenticate([]byte(password), user.Password, user.Salt).Return(nil)
	mockAuthEngine.EXPECT().GenerateToken(user.ID).Return(token, nil).Times(1)

	tkn, encryptSalt, err := userService.Login(ctx, login, password)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, encryptSalt)
	assert.Equal(t, token, tkn)
	assert.Equal(t, user.EncryptSalt, encryptSalt)
}

func TestUserService_Login_FailToLoginWithWrongPassword(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUow := mocks.NewMockUnitOfWork(ctrl)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockAuthEngine := mocks.NewMockEngine(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	userService := createUserService(mockUow, mockAuthService, mockAuthEngine)

	login := "dima"
	wrongPassword := "qwerty1"
	password := "qwerty"
	user := &domain.User{
		ID:    1,
		Login: login,

		Password:    []byte(password),
		Salt:        []byte("salt"),
		EncryptSalt: []byte("encrypt_salt"),
	}
	token := "token"
	mockUow.EXPECT().UserRepository().Return(mockRepo).AnyTimes()
	mockRepo.EXPECT().Exist(ctx, login).Return(true, nil)
	mockRepo.EXPECT().Get(ctx, login).Return(user, nil).Times(1)
	mockAuthService.EXPECT().Authenticate([]byte(wrongPassword), user.Password, user.Salt).Return(auth.ErrInvalidPassword)
	mockAuthEngine.EXPECT().GenerateToken(user.ID).Return(token, nil).Times(0)

	tkn, encryptSalt, err := userService.Login(ctx, login, wrongPassword)

	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrInvalidPassword)
	assert.Empty(t, tkn)
	assert.Nil(t, encryptSalt)
}

func TestUserService_Login_FailToLoginWithWrongLogin(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUow := mocks.NewMockUnitOfWork(ctrl)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockAuthEngine := mocks.NewMockEngine(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	userService := createUserService(mockUow, mockAuthService, mockAuthEngine)

	wrongLogin := "dima1"
	login := "dima"
	password := "qwerty"
	user := &domain.User{
		ID:    1,
		Login: login,

		Password:    []byte(password),
		Salt:        []byte("salt"),
		EncryptSalt: []byte("encrypt_salt"),
	}
	token := "token"
	mockUow.EXPECT().UserRepository().Return(mockRepo).AnyTimes()
	mockRepo.EXPECT().Exist(ctx, wrongLogin).Return(false, nil)
	mockRepo.EXPECT().Get(ctx, wrongLogin).Return(user, nil).Times(0)
	mockAuthService.EXPECT().Authenticate([]byte(password), user.Password, user.Salt).Return(nil).Times(0)
	mockAuthEngine.EXPECT().GenerateToken(user.ID).Return(token, nil).Times(0)

	tkn, encryptSalt, err := userService.Login(ctx, wrongLogin, password)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Empty(t, tkn)
	assert.Nil(t, encryptSalt)
}

func createUserService(work domain.UnitOfWork, authService auth.AuthService, engine auth.Engine) *UserService {
	userService := NewUserService(work, authService, engine)
	return userService
}
