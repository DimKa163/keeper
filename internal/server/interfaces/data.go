package interfaces

import (
	"context"
	"errors"
	"github.com/DimKa163/keeper/internal/pb"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/shared/auth"
	"github.com/DimKa163/keeper/internal/server/shared/logging"
	"github.com/DimKa163/keeper/internal/server/usecase"
	"github.com/DimKa163/keeper/internal/shared"
	"github.com/beevik/guid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
)

type DataServer struct {
	app *usecase.DataService
	pb.UnimplementedSyncDataServer
}

func NewDataServer(app *usecase.DataService) *DataServer {
	return &DataServer{app: app}
}

func (ds *DataServer) Bind(server *grpc.Server) {
	pb.RegisterSyncDataServer(server, ds)
}

func (ds *DataServer) Push(ctx context.Context, request *pb.PushRequest) (*pb.PushResponse, error) {
	var response pb.PushResponse
	req, err := toPushUnaryRequest(ctx, request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = ds.app.Push(ctx, req)
	if err != nil {
		if errors.Is(err, usecase.ErrDataConflict) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	response.SetSuccess(true)
	return &response, nil
}

func (ds *DataServer) PushStream(stream pb.SyncData_PushStreamServer) error {
	ctx := stream.Context()
	for {
		logger := logging.Logger(ctx)
		var resp pb.PushResponse
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Debug("finish push stream request")
				return stream.SendAndClose(&resp)
			}
			return err
		}

		switch req.GetType() {
		case pb.ChunkType_FilePart:
			request, err := toPushChunkRequest(req)
			if err != nil {
				logger.Error("validation failed", zap.Error(err))
				return status.Error(codes.InvalidArgument, err.Error())
			}
			if err := ds.app.HandleChunk(ctx, request); err != nil {
				logger.Error("failed pushing req", zap.Error(err))
				return status.Error(codes.Internal, err.Error())
			}
		case pb.ChunkType_EndData:
			request, err := toHandleCloser(req)
			if err != nil {
				logger.Error("validation failed", zap.Error(err))
				return status.Error(codes.InvalidArgument, err.Error())
			}
			if err = ds.app.Commit(ctx, request); err != nil {
				logger.Error("failed pushing req", zap.Error(err))
				return status.Error(codes.Internal, err.Error())
			}
		}

	}
}

func (ds *DataServer) Pull(ctx context.Context, request *pb.PullRequest) (*pb.PullResponse, error) {
	var response pb.PullResponse
	data, v, err := ds.app.Poll(ctx, request.GetSince())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response.SetVersion(v)
	secrets := make([]*pb.Secret, len(data))
	for i, item := range data {
		secret := toSecret(item)
		secrets[i] = secret
	}
	response.SetSecrets(secrets)
	return &response, nil
}
func (ds *DataServer) PullStream(in *pb.PullStreamRequest, stream pb.SyncData_PullStreamServer) error {
	ctx := stream.Context()
	logger := logging.Logger(ctx)
	reader, err := ds.app.OpenFile(in.GetId(), in.GetVersion())
	if err != nil {
		chunk := pb.Chunk{}
		chunk.SetType(pb.ChunkType_ErrData)
		chunk.SetVersion(in.GetVersion())
		chunk.SetId(in.GetId())
		if err := stream.Send(&chunk); err != nil {
			return err
		}
		return nil
	}
	defer func(file io.ReadCloser) {
		err = file.Close()
		if err != nil {
			logger.Warn("failed to close file", zap.Error(err))
		}
	}(reader)
	buffer := make([]byte, shared.MB)
	for {
		var chunk pb.Chunk
		var n int
		n, err = reader.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			logger.Error("failed to read data", zap.Error(err))
			return err
		}
		chunk.SetType(pb.ChunkType_FilePart)
		chunk.SetVersion(in.GetVersion())
		chunk.SetId(in.GetId())
		chunk.SetBuffer(buffer[:n])
		if err := stream.Send(&chunk); err != nil {
			logger.Warn("failed to send chunk", zap.Error(err))
			return err
		}
	}
	return nil
	//ctx := stream.Context()
	//logger := logging.Logger(ctx)
	//data, version, err := ds.app.pull(ctx, in.GetSince())
	//if err != nil {
	//	return status.Error(codes.Internal, err.Error())
	//}
	//loggerSg := logger.Sugar()
	//loggerSg.Infof("changes: %d; last version: %d", len(data), version)
	//for _, d := range data {
	//	var poll pb.pull
	//	poll.SetServerVersion(version)
	//	md := toMetadata(d)
	//	dt := toData(d)
	//	if d.Deleted {
	//		poll.SetMetadata(md)
	//		poll.SetData(dt)
	//		poll.SetType(pb.RequestType_Default)
	//		if err = stream.Send(&poll); err != nil {
	//			return status.Error(codes.Internal, err.Error())
	//		}
	//		continue
	//	}
	//	if d.BigData {
	//		poll.SetType(pb.RequestType_StartData)
	//		poll.SetMetadata(md)
	//		if err = stream.Send(&poll); err != nil {
	//			return status.Error(codes.Internal, err.Error())
	//		}
	//		file, err := ds.app.OpenFile(d.Path)
	//		if err != nil {
	//			if !os.IsNotExist(err) {
	//				return status.Error(codes.Internal, err.Error())
	//			}
	//			poll = pb.pull{}
	//			poll.SetServerVersion(version)
	//			poll.SetType(pb.RequestType_ErrData)
	//			poll.SetMetadata(md)
	//			continue
	//		}
	//		defer func(file io.ReadCloser) {
	//			err := file.Close()
	//			if err != nil {
	//				logger.Warn("failed to close file", zap.Error(err))
	//			}
	//		}(file)
	//		buffer := make([]byte, shared.MB)
	//		for {
	//			poll = pb.pull{}
	//			poll.SetServerVersion(version)
	//			poll.SetType(pb.RequestType_FilePart)
	//			poll.SetMetadata(md)
	//
	//			n, err := file.Read(buffer)
	//			if err != nil {
	//				if err == io.EOF {
	//					break
	//				}
	//				return status.Error(codes.Internal, err.Error())
	//			}
	//			if n == 0 {
	//				break
	//			}
	//			poll.SetChunk(buffer[:n])
	//			if err = stream.Send(&poll); err != nil {
	//				return status.Error(codes.Internal, err.Error())
	//			}
	//		}
	//		poll = pb.pull{}
	//		poll.SetServerVersion(version)
	//		poll.SetType(pb.RequestType_EndData)
	//		poll.SetMetadata(md)
	//		poll.SetData(dt)
	//		if err = stream.Send(&poll); err != nil {
	//
	//			logger.Error(err.Error())
	//			return status.Error(codes.Internal, err.Error())
	//		}
	//
	//	} else {
	//		poll.SetMetadata(md)
	//		poll.SetData(dt)
	//		poll.SetType(pb.RequestType_Default)
	//		if err = stream.Send(&poll); err != nil {
	//			return status.Error(codes.Internal, err.Error())
	//		}
	//	}
	//}
	//return nil
}

func toPushUnaryRequest(ctx context.Context, push *pb.PushRequest) (*usecase.PushUnaryRequest, error) {
	secrets := push.GetSecrets()
	dt := usecase.PushUnaryRequest{
		Secrets: make([]*usecase.Secret, len(secrets)),
		Force:   push.GetForce(),
	}
	for i, secret := range secrets {
		var sc usecase.Secret
		id, err := guid.ParseString(secret.GetId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		user, err := auth.User(ctx)
		if err != nil {
			return nil, err
		}
		sc.ID = *id
		sc.ModifiedAt = secret.GetModifiedAt().AsTime()
		sc.BigData = secret.GetIsBig()
		sc.Payload = secret.GetData()
		sc.UserID = user
		sc.Dek = secret.GetDek()
		sc.Version = secret.GetVersion()
		sc.Deleted = secret.GetDeleted()
		switch secret.GetType() {
		case pb.Secret_LoginPass:
			sc.Type = domain.LoginPassType
		case pb.Secret_Text:
			sc.Type = domain.TextType
		case pb.Secret_BankCard:
			sc.Type = domain.BankCardType
		case pb.Secret_Other:
			sc.Type = domain.OtherType
		}
		dt.Secrets[i] = &sc
	}

	return &dt, nil
}

func toPushChunkRequest(chunk *pb.Chunk) (*usecase.PushChunkRequest, error) {
	var req usecase.PushChunkRequest
	id, err := guid.ParseString(chunk.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	req.ID = *id
	req.Buffer = chunk.GetBuffer()
	return &req, nil
}

func toHandleCloser(chunk *pb.Chunk) (*usecase.PushFileCloseRequest, error) {
	var req usecase.PushFileCloseRequest
	secret := chunk.GetSecret()
	id, err := guid.ParseString(chunk.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	req.ID = *id
	req.Dek = secret.GetDek()
	req.Payload = secret.GetData()
	req.ModifiedAt = secret.GetModifiedAt().AsTime()
	return &req, nil
}

func toSecret(data *domain.Data) *pb.Secret {
	var secret pb.Secret
	secret.SetId(data.ID.String())
	secret.SetModifiedAt(timestamppb.New(data.ModifiedAt))
	secret.SetIsBig(data.BigData)
	secret.SetVersion(int32(data.Version))
	secret.SetDek(data.Dek)
	secret.SetData(data.Payload)
	secret.SetDeleted(data.Deleted)
	switch data.Type {
	case domain.LoginPassType:
		secret.SetType(pb.Secret_LoginPass)
	case domain.TextType:
		secret.SetType(pb.Secret_Text)
	case domain.BankCardType:
		secret.SetType(pb.Secret_BankCard)
	case domain.OtherType:
		secret.SetType(pb.Secret_Other)
	}
	return &secret
}
