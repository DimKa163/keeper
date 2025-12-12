package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	server2 "github.com/DimKa163/keeper/app/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/DimKa163/keeper/internal/pb"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
)

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()
	container, serv, err := run(ctx, t, func(s *services) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(ctx)
	defer serv.DBPool.Close()
	defer serv.Shutdown(ctx)

	client := serv.UsersClient

	req := pb.User{}

	req.SetLogin("dima")
	req.SetPassword("123")

	resp, err := client.Register(ctx, &req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.HasToken())
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()

	login := "dima"
	password := "123"

	container, serv, err := run(ctx, t, func(s *services) error {
		t.Log("generate data")
		pwd, salt, err := s.AuthService.GenerateHash([]byte(password))
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
		return s.UnitOfWork.UserRepository().Insert(ctx, domain.NewUser(login, pwd, salt))
	})
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(ctx)
	defer serv.DBPool.Close()
	defer serv.Shutdown(ctx)

	client := serv.UsersClient

	req := pb.User{}

	req.SetLogin(login)
	req.SetPassword(password)

	resp, err := client.Login(ctx, &req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.HasToken())
}

//func TestDataService_Push(t *testing.T) {
//	ctx := context.Background()
//	container, serv, err := run(ctx, t, func(s *services) error {
//		user, _ := s.UnitOfWork.UserRepository().Get(ctx, "root")
//		rep := s.UnitOfWork.SyncStateRepository()
//		state, _ := rep.Get(ctx, "Secret", user.ID)
//		state.Value = 2
//		_ = rep.Update(ctx, state)
//		return nil
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer container.Terminate(ctx)
//	defer serv.DBPool.Close()
//	defer serv.Shutdown(ctx)
//	user, _ := serv.UnitOfWork.UserRepository().Get(ctx, "root")
//	client := serv.DataClient
//	id := *guid.New()
//	// генерим ерунду так как серверу пофиг что там внутри
//	dek, dekNonce := make([]byte, 32), make([]byte, 16)
//	_, _ = rand.Read(dek)
//	_, _ = rand.Read(dekNonce)
//	dt, dtNonce := make([]byte, 1024*1024*50), make([]byte, 16)
//	_, _ = rand.Read(dt)
//	_, _ = rand.Read(dtNonce)
//	str, err := client.PushStream(ctx)
//	if err != nil {
//		t.Fatal(err)
//	}
//	var push pb.PushRequest
//	var data pb.Secret
//	data.SetId(id.String())
//	data.SetDek(dek)
//	data.SetType(pb.Secret_Other)
//	data.SetIsBig(true)
//	data.SetVersion(1)
//	push.SetSecrets(&data)
//	push.SetType(pb.RequestType_StartData)
//	if err := str.Send(&push); err != nil {
//		t.Fatal(err)
//	}
//	const chunkSize = 1 * 1024 * 1024 // 1
//	for i := 0; i < len(dt); i += chunkSize {
//		end := i + chunkSize
//		push = pb.Push{}
//		data = pb.Secret{}
//		data.SetId(id.String())
//		push.SetData(&data)
//		push.SetType(pb.RequestType_FilePart)
//		push.SetChunk(dt[i:end])
//		if err := str.Send(&push); err != nil {
//			t.Fatal(err)
//		}
//	}
//	push = pb.Push{}
//	data = pb.Secret{}
//	data.SetId(id.String())
//	push.SetData(&data)
//	push.SetType(pb.RequestType_EndData)
//	if err := str.Send(&push); err != nil {
//		t.Fatal(err)
//	}
//
//	rep := serv.UnitOfWork.SyncStateRepository()
//	state, _ := rep.Get(ctx, "Secret", user.ID)
//	assert.Equal(t, state.Value, int64(3))
//}

func run(ctx context.Context, t *testing.T, beforeFn func(pool *services) error) (testcontainers.Container, *services, error) {
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
		return nil, nil, err
	}

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/%s?sslmode=disable", host, port.Port(), dbName)
	t.Logf("postgres started at: %s", dsn)

	serv := &services{}
	addr := ":3200"

	if err = configureServer(t, serv, &server2.Config{
		Addr:            addr,
		Database:        dsn,
		FilePath:        "E:\\Secret",
		Secret:          "secret",
		TokenExpiration: 5000,
		SaltLength:      16,
		Iterations:      4,
		Parallelism:     2,
		Memory:          64 * 1024,
		KeyLength:       32,
	}); err != nil {
		return nil, nil, err
	}

	login := "root"
	password := "root"

	if err = createRootUser(t, serv, login, password); err != nil {
		return nil, nil, err
	}

	if err = configureClient(t, serv, addr, login, password); err != nil {
		return nil, nil, err
	}

	if err := beforeFn(serv); err != nil {
		return nil, nil, err
	}

	return pgContainer, serv, nil
}

func configureServer(t *testing.T, serv *services, config *server2.Config) error {
	t.Logf("configure application server on %s", config.Addr)
	srv, err := server2.NewServer(config, "", "", "")
	if err != nil {
		return err
	}
	if err := srv.AddServices(); err != nil {
		t.Fatal(err)
		return err
	}
	srv.Map()
	if err := srv.AddLogging(); err != nil {
		return err
	}
	if err := srv.MigrateFrom("../internal/server/migrations"); err != nil {
		return err
	}
	go func() {
		t.Log("starting server")
		_ = srv.Run()
	}()
	serv.Server = srv
	return nil
}

func createRootUser(t *testing.T, serv *services, login, pass string) error {
	t.Log("create root user")
	pwd, salt, err := serv.AuthService.GenerateHash([]byte(pass))
	if err != nil {
		return err
	}
	err = serv.UnitOfWork.UserRepository().Insert(context.Background(), domain.NewUser(login, pwd, salt))
	if err != nil {
		return err
	}
	return nil
}

func configureClient(t *testing.T, serv *services, addr, login, pass string) error {
	t.Log("configure client")
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	serv.UsersClient = pb.NewUsersClient(conn)
	serv.interceptor = newInterceptor(serv.UsersClient, login, pass)
	serv.streamInterceptor = newStreamInterceptor(serv.UsersClient, login, pass)
	protectedConn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithChainUnaryInterceptor(serv.interceptor.Handle()),
		grpc.WithChainStreamInterceptor(serv.streamInterceptor.Handle()))
	serv.DataClient = pb.NewSyncClient(protectedConn)
	return err
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
					if err != nil {
						return err
					}
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

type services struct {
	*server2.Server
	interceptor       *unaryIdentifyInterceptor
	streamInterceptor *streamIdentifyInterceptor
	UsersClient       pb.UsersClient
	DataClient        pb.SyncClient
}
