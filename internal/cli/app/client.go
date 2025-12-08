package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/DimKa163/keeper/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrServerUnavailable = errors.New("server is unavailable")
)

type RemoteClient struct {
	addr  string
	login string
	pass  string
	pb.HealthServiceClient
	pb.UsersClient
	pb.SyncClient
}

func NewRemoteClient(addr string, login string, pass string) (*RemoteClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	usersClient := pb.NewUsersClient(conn)
	protectedConn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(newInterceptor(usersClient, login, pass).Handle()),
		grpc.WithChainStreamInterceptor(newStreamInterceptor(usersClient, login, pass).Handle()))
	if err != nil {
		return nil, err
	}
	return &RemoteClient{
		login:               login,
		pass:                pass,
		addr:                addr,
		HealthServiceClient: pb.NewHealthServiceClient(conn),
		UsersClient:         usersClient,
		SyncClient:          pb.NewSyncClient(protectedConn),
	}, nil
}

func (rm *RemoteClient) IsHealthy(ctx context.Context) error {
	res, err := rm.Check(ctx, &pb.HealthCheckRequest{})
	if err != nil {
		code, ok := status.FromError(err)
		if ok {
			if code.Code() == codes.Unavailable {
				return ErrServerUnavailable
			}
		}
		return err
	}
	if res.GetState() != pb.ServerState_Healthy {
		return errors.New("remote server is not serving")
	}
	return nil
}

func (rm *RemoteClient) TryToAuthenticate(ctx context.Context) error {
	var us pb.User
	us.SetLogin(rm.login)
	us.SetPassword(rm.pass)
	_, err := rm.Login(ctx, &us)
	if err != nil {
		code, ok := status.FromError(err)
		if !ok {
			return err
		}
		if code.Code() != codes.Unauthenticated {
			return err
		}
		fmt.Println("authentification failed. trying create a new user")
		if _, err = rm.Register(ctx, &us); err != nil {
			return err
		}
		fmt.Println("✅ new user was created successfully")
		return nil
	}
	fmt.Println("✅ authentification succeeded.")
	return nil
}

type unaryIdentifyInterceptor struct {
	users    pb.UsersClient
	token    string
	username string
	userpass string
}

func newInterceptor(users pb.UsersClient, username, userpass string) *unaryIdentifyInterceptor {
	return &unaryIdentifyInterceptor{
		users:    users,
		username: username,
		userpass: userpass,
	}
}

func (h *unaryIdentifyInterceptor) Handle() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		if h.token == "" {
			h.token, err = h.login(ctx)
			if err != nil {
				return err
			}
		}
		md := metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", h.token)})
		err = invoker(metadata.NewOutgoingContext(ctx, md), method, req, reply, cc, opts...)
		if err != nil {
			if e, ok := status.FromError(err); ok {
				if e.Code() == codes.Unauthenticated {
					h.token, err = h.login(ctx)
					//md = metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", h.token)})
					//ctx = metadata.NewOutgoingContext(ctx, md)
					err = invoker(ctx, method, req, reply, cc, opts...)
				}
			}
		}
		return err
	}
}

func (h *unaryIdentifyInterceptor) login(ctx context.Context) (string, error) {
	var us pb.User
	us.SetLogin(h.username)
	us.SetPassword(h.userpass)
	t, err := h.users.Login(ctx, &us)
	if err != nil {
		return "", err
	}
	return t.GetToken(), nil
}

type streamIdentifyInterceptor struct {
	users    pb.UsersClient
	token    string
	username string
	userpass string
}

func newStreamInterceptor(users pb.UsersClient, username, userpass string) *streamIdentifyInterceptor {
	return &streamIdentifyInterceptor{
		users:    users,
		username: username,
		userpass: userpass,
	}
}

func (h *streamIdentifyInterceptor) Handle() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		var err error
		if h.token == "" {
			h.token, err = h.login(ctx)
			if err != nil {
				return nil, err
			}
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", h.token)

		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}

		return stream, nil
	}
}

func (h *streamIdentifyInterceptor) login(ctx context.Context) (string, error) {
	var us pb.User
	us.SetLogin(h.username)
	us.SetPassword(h.userpass)
	t, err := h.users.Login(ctx, &us)
	if err != nil {
		return "", err
	}
	return t.GetToken(), nil
}
