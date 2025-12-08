// Package server
package server

import (
	"context"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimKa163/keeper/internal/server/infrastructure/data"
	"github.com/DimKa163/keeper/internal/shared"

	"github.com/DimKa163/keeper/internal/server/interfaces"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/domain/auth"
	"github.com/DimKa163/keeper/internal/server/infrastructure/persistence"
	"github.com/DimKa163/keeper/internal/server/infrastructure/security"
	"github.com/DimKa163/keeper/internal/server/shared/logging"
	"github.com/DimKa163/keeper/internal/server/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

type ServiceContainer struct {
	DBPool        *pgxpool.Pool
	UnitOfWork    domain.UnitOfWork
	AuthService   auth.AuthService
	AuthEngine    auth.Engine
	UserService   domain.UserService
	SyncService   *usecase.SyncService
	UserRPCServer *interfaces.UsersServer
	SyncRPCServer *interfaces.SyncServer
	HealthServer  *interfaces.HealthService
}

type Server struct {
	listener net.Listener
	Version  string
	Commit   string
	Date     string
	*Config
	*ServiceContainer
	ServerImpl
}

func NewServer(config *Config, version, commit, date string) (*Server, error) {
	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return nil, err
	}
	return &Server{
		Config:           config,
		ServiceContainer: &ServiceContainer{},
		listener:         listener,
		Version:          version,
		Commit:           commit,
		Date:             date,
	}, nil
}

func (server *Server) AddServices() error {
	var err error
	server.DBPool, err = addPgPool(server.Database)
	if err != nil {
		return err
	}
	server.HealthServer = interfaces.NewHealthService(server.DBPool)
	server.AuthEngine = addAuthEngine(server.Config)
	server.ServerImpl = NewGRPCServer(server.listener, addGrpcServer(server.ServiceContainer), server.ServiceContainer)
	server.UnitOfWork = addUnitOfWork(server.DBPool)
	server.AuthService = addAuthService(server.Config)
	server.UserService = addUserService(server.UnitOfWork, server.AuthService, server.AuthEngine)
	server.UserRPCServer = interfaces.NewUserServer(server.UserService)
	server.SyncService = usecase.NewSyncService(server.UnitOfWork, data.NewFileProvider(shared.NewFileProvider(server.FilePath)))
	server.SyncRPCServer = interfaces.NewSyncServer(server.SyncService)
	return nil
}

func (server *Server) AddLogging() error {
	return logging.InitializeLogging(&logging.LogConfiguration{
		Builders: map[string]logging.CoreBuilder{
			"console": logging.NewConsoleBuilder(zap.NewDevelopmentEncoderConfig(), zapcore.DebugLevel),
		},
	})
}

func (server *Server) Map() {
	server.ServerImpl.Map()
}

func (server *Server) Migrate() error {
	return persistence.Migrate(server.DBPool, "./internal/server/migrations")
}

func (server *Server) MigrateFrom(path string) error {
	return persistence.Migrate(server.DBPool, path)
}

func (server *Server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	logger := logging.Logger(ctx).Sugar()
	logger.Infof("version: %s; commit: %s; data: %s", ifNan(server.Version), ifNan(server.Commit), ifNan(server.Date))
	go func() {
		<-ctx.Done()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = server.ServerImpl.Shutdown(timeoutCtx)
	}()
	return server.ListenAndServe()
}

func addPgPool(database string) (*pgxpool.Pool, error) {
	pg, err := pgxpool.New(context.Background(), database)
	if err != nil {
		return nil, err
	}
	return pg, nil
}

func addUnitOfWork(pool *pgxpool.Pool) domain.UnitOfWork {
	return persistence.NewUnitOfWork(pool)
}
func addGrpcServer(container *ServiceContainer) *grpc.Server {
	chain := make([]grpc.UnaryServerInterceptor, 0)
	chain = append(chain, interfaces.UnaryLoggingInterceptor())
	skip := make(map[string]bool)
	skip["/go.Users/Login"] = true
	skip["/go.Users/Register"] = true
	skip["/go.HealthService/Check"] = true
	chain = append(chain, interfaces.UnaryIdentifyInterceptor(container.AuthEngine, skip))
	streamChain := make([]grpc.StreamServerInterceptor, 0)
	streamChain = append(streamChain, interfaces.StreamIdentifyInterceptor(container.AuthEngine))
	return grpc.NewServer(grpc.ChainUnaryInterceptor(chain...), grpc.ChainStreamInterceptor(streamChain...))
}

func addAuthService(config *Config) auth.AuthService {
	return auth.NewAuthService(&auth.ArgonConfig{
		Memory:      uint32(config.Memory),
		Iterations:  uint32(config.Iterations),
		Parallelism: uint32(config.Parallelism),
		SaltLength:  uint32(config.SaltLength),
		KeyLength:   uint32(config.KeyLength),
	})
}

func addAuthEngine(config *Config) auth.Engine {
	return security.NewJWTEngine(&security.JWTConfig{
		SecretKey:       []byte(config.Secret),
		TokenExpiration: time.Duration(config.TokenExpiration) * time.Second,
	})
}

func addUserService(unitOfWork domain.UnitOfWork, authService auth.AuthService, engine auth.Engine) domain.UserService {
	return usecase.NewUserService(unitOfWork, authService, engine)
}

type ServerImpl interface {
	ListenAndServe() error
	Map()
	Shutdown(ctx context.Context) error
}

type GRPCServer struct {
	services *ServiceContainer
	listener net.Listener
	*grpc.Server
}

func NewGRPCServer(listener net.Listener, server *grpc.Server, services *ServiceContainer) *GRPCServer {
	return &GRPCServer{
		Server:   server,
		listener: listener,
		services: services,
	}
}
func (gs *GRPCServer) ListenAndServe() error {
	logger := logging.GetLogger()
	loggerSugar := logger.Sugar()
	loggerSugar.Infof("Listening on %s", gs.listener.Addr())
	return gs.Serve(gs.listener)
}

func (gs *GRPCServer) Map() {
	gs.services.HealthServer.Bind(gs.Server)
	gs.services.UserRPCServer.Bind(gs.Server)
	gs.services.SyncRPCServer.Bind(gs.Server)
}

func (gs *GRPCServer) Shutdown(ctx context.Context) error {
	logger := logging.Logger(ctx)
	gs.GracefulStop()
	logger.Info("server shutdown gracefully")
	return nil
}

func ifNan(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}
