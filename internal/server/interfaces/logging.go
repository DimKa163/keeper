package interfaces

import (
	"context"
	"time"

	"github.com/DimKa163/keeper/internal/server/shared/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logg := logging.Logger(ctx)
		logg = logg.With(zap.String("method", info.FullMethod))
		ctx = logging.SetLogger(ctx, logg)
		logg.Info("got incoming grpc request")
		startTime := time.Now()
		resp, err := handler(ctx, req)
		elapsed := time.Since(startTime)
		if err != nil {
			logg.Warn("Processed with error", zap.String("error", err.Error()))
		}
		logg.Info("grpc request processed", zap.Duration("elapsed", elapsed))
		return resp, err
	}
}
