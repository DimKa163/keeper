package interfaces

import (
	"context"
	"errors"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/interfaces/pb"
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

	token, salt, err := us.app.Login(ctx, in.GetLogin(), in.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response.SetToken(token)
	response.SetEncryptedSalt(salt)

	return &response, nil
}

func (us *UsersServer) Register(ctx context.Context, in *pb.User) (*pb.UserResponse, error) {
	var response pb.UserResponse
	if err := validate(in); err != nil {
		return nil, err
	}
	token, salt, err := us.app.Register(ctx, in.GetLogin(), in.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response.SetToken(token)
	response.SetEncryptedSalt(salt)
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
