package server

import (
	"context"
	"github.com/DimKa163/keeper/internal/server/interfaces"
	"net"
	"time"

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
	DBPool       *pgxpool.Pool
	UnitOfWork   domain.UnitOfWork
	AuthService  auth.AuthService
	UserService  domain.UserService
	UserGcServer *interfaces.UsersServer
}

type Server struct {
	listener net.Listener
	*Config
	*ServiceContainer
	ServerImpl
}

func NewServer(config *Config) (*Server, error) {
	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return nil, err
	}
	return &Server{
		Config:           config,
		ServiceContainer: &ServiceContainer{},
		listener:         listener,
	}, nil
}

func (server *Server) AddServices() error {
	var err error
	server.ServerImpl = NewGRPCServer(server.listener, addGrpcServer(), server.ServiceContainer)
	server.DBPool, err = addPgPool(server.Database)
	if err != nil {
		return err
	}
	server.UnitOfWork = addUnitOfWork(server.DBPool)
	server.AuthService = addAuthService(server.Config)
	server.UserService = addUserService(server.UnitOfWork, server.AuthService, addAuthEngine(server.Config))
	server.UserGcServer = interfaces.NewUserServer(server.UserService)
	return nil
}

func (server *Server) AddLogging() error {
	return logging.InitializeLogging(&logging.LogConfiguration{
		Builders: map[string]logging.CoreBuilder{
			"file":    logging.NewFileBuilder("D:\\logs\\keeper.log", zap.NewProductionEncoderConfig(), zapcore.InfoLevel),
			"console": logging.NewConsoleBuilder(zap.NewDevelopmentEncoderConfig(), zapcore.DebugLevel),
		},
	})
}

func (server *Server) Map() {
	server.ServerImpl.Map()
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
func addGrpcServer() *grpc.Server {
	return grpc.NewServer(grpc.ChainUnaryInterceptor())
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
	gs.services.UserGcServer.Bind(gs.Server)
}

func (gs *GRPCServer) Shutdown(ctx context.Context) error {
	logger := logging.Logger(ctx)
	gs.GracefulStop()
	logger.Info("server shutdown gracefully")
	return nil
}
