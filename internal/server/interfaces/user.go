package interfaces

import (
	"context"
	"errors"

	"github.com/DimKa163/keeper/internal/server/usecase"

	"github.com/DimKa163/keeper/internal/pb"

	"github.com/DimKa163/keeper/internal/server/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UsersServer struct {
	app domain.UserService
	pb.UnimplementedUsersServer
}

func NewUserServer(app domain.UserService) *UsersServer {
	return &UsersServer{app: app}
}

func (us *UsersServer) Bind(server *grpc.Server) {
	pb.RegisterUsersServer(server, us)
}

func (us *UsersServer) Login(ctx context.Context, in *pb.User) (*pb.UserResponse, error) {
	var response pb.UserResponse
	if err := validate(in); err != nil {
		return nil, err
	}

	token, err := us.app.Login(ctx, in.GetLogin(), in.GetPassword())
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	response.SetToken(token)

	return &response, nil
}

func (us *UsersServer) Register(ctx context.Context, in *pb.User) (*pb.UserResponse, error) {
	var response pb.UserResponse
	if err := validate(in); err != nil {
		return nil, err
	}
	token, err := us.app.Register(ctx, in.GetLogin(), in.GetPassword())
	if err != nil {
		if errors.Is(err, usecase.ErrLoginAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	response.SetToken(token)
	return &response, nil
}

func validate(in *pb.User) error {
	errs := make([]error, 0)
	if !in.HasLogin() {
		errs[0] = errors.New("login required")
	}
	if !in.HasPassword() {
		errs[1] = errors.New("password required")
	}
	if len(errs) != 0 {
		return status.Error(codes.InvalidArgument, errors.Join(errs...).Error())
	}
	return nil
}
