package interfaces

import (
	"context"
	"github.com/DimKa163/keeper/internal/pb"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type HealthService struct {
	db *pgxpool.Pool
	pb.UnimplementedHealthServiceServer
}

func NewHealthService(db *pgxpool.Pool) *HealthService {
	return &HealthService{db: db}
}

func (hs *HealthService) Bind(server *grpc.Server) {
	pb.RegisterHealthServiceServer(server, hs)
}

func (hs *HealthService) Check(ctx context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	var resp pb.HealthCheckResponse
	if err := hs.db.Ping(ctx); err != nil {
		resp.SetState(pb.ServerState_NotHealthy)
		return &resp, nil
	}
	resp.SetState(pb.ServerState_Healthy)
	return &resp, nil
}
