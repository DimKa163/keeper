package integration_test

import (
	"context"
	"fmt"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/interfaces"
	"github.com/DimKa163/keeper/internal/server/interfaces/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"

	"github.com/DimKa163/keeper/internal/server/domain/auth"
	"github.com/DimKa163/keeper/internal/server/infrastructure/persistence"
	"github.com/DimKa163/keeper/internal/server/infrastructure/security"
	"github.com/DimKa163/keeper/internal/server/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()
	container, server, serv, err := run(ctx, t, func(s *services) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(ctx)
	defer server.Stop()
	defer serv.Pool.Close()

	client := serv.UsersClient

	req := pb.User{}

	req.SetLogin("dima")
	req.SetPassword("123")

	resp, err := client.Register(ctx, &req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.HasToken())
	assert.True(t, resp.HasEncryptedSalt())
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()

	login := "dima"
	password := "123"

	container, server, serv, err := run(ctx, t, func(s *services) error {
		pwd, salt, err := s.AuthService.GenerateHash([]byte(password))
		if err != nil {
			return err
		}
		encSalt, err := s.AuthService.GenerateSalt()
		if err != nil {
			return err
		}
		return s.Uow.UserRepository().Insert(ctx, domain.NewUser(login, pwd, salt, encSalt))
	})
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(ctx)
	defer server.Stop()
	defer serv.Pool.Close()

	client := serv.UsersClient

	req := pb.User{}

	req.SetLogin(login)
	req.SetPassword(password)

	resp, err := client.Login(ctx, &req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.HasToken())
	assert.True(t, resp.HasEncryptedSalt())
}

func run(ctx context.Context, t *testing.T, beforeFn func(pool *services) error) (testcontainers.Container, *grpc.Server, *services, error) {
	dbName := "keeperDb"
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       dbName,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").
			WithStartupTimeout(90 * time.Second),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, nil, nil, err
	}
	addr := ":3300"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, nil, err
	}
	server := grpc.NewServer(grpc.ChainUnaryInterceptor())
	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/%s?sslmode=disable", host, port.Port(), dbName)
	t.Logf("Started postgres at: %s", dsn)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, nil, err
	}

	serv := &services{}
	serv.Pool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := persistence.Migrate(serv.Pool, "../migrations"); err != nil {
		return nil, nil, nil, err
	}
	serv.Uow = persistence.NewUnitOfWork(serv.Pool)

	serv.Engine = security.NewJWTEngine(&security.JWTConfig{
		SecretKey:       []byte("secret"),
		TokenExpiration: 5 * time.Minute,
	})

	serv.AuthService = auth.NewAuthService(&auth.ArgonConfig{
		SaltLength:  16,
		Iterations:  4,
		Parallelism: 2,
		Memory:      64 * 1024,
		KeyLength:   32,
	})

	serv.UserService = usecase.NewUserService(serv.Uow, serv.AuthService, serv.Engine)
	serv.UserServer = interfaces.NewUserServer(serv.UserService)
	serv.UserServer.Bind(server)
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Error("Failed to start server:", err)
			return
		}
	}()

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, err
	}
	serv.UsersClient = pb.NewUsersClient(conn)

	if err := beforeFn(serv); err != nil {
		return nil, nil, nil, err
	}

	return pgContainer, server, serv, nil
}

type services struct {
	Pool        *pgxpool.Pool
	Uow         domain.UnitOfWork
	Engine      auth.Engine
	AuthService auth.AuthService
	UserService *usecase.UserService
	UserServer  *interfaces.UsersServer
	UsersClient pb.UsersClient
}
