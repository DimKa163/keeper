package interfaces

import (
	"context"

	"github.com/DimKa163/keeper/internal/server/domain/auth"
	sh "github.com/DimKa163/keeper/internal/server/shared/auth"
	"github.com/beevik/guid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryIdentifyInterceptor(engine auth.Engine, skip map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := skip[info.FullMethod]; ok {
			return handler(ctx, req)
		}
		mtdata, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata not found in context")
		}
		val := mtdata.Get("authorization")
		if len(val) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization not found in context")
		}
		cl, err := engine.ReadToken(val[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "authorization expired")
		}
		id, err := guid.ParseString(cl.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, "unrecognized user id")
		}
		ctx = sh.SetUser(ctx, *id)
		return handler(ctx, req)
	}
}

func StreamIdentifyInterceptor(engine auth.Engine) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}
		if !ok {
			return status.Error(codes.Unauthenticated, "metadata not found in context")
		}
		val := md.Get("authorization")
		if len(val) == 0 {
			return status.Error(codes.Unauthenticated, "authorization not found in context")
		}
		cl, err := engine.ReadToken(val[0])
		if err != nil {
			return status.Error(codes.Unauthenticated, "authorization expired")
		}
		id, err := guid.ParseString(cl.ID)
		if err != nil {
			return status.Error(codes.Internal, "unrecognized user id")
		}
		ctx = sh.SetUser(ctx, *id)
		return handler(srv, &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		})
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *wrappedServerStream) Context() context.Context {
	return ss.ctx
}
