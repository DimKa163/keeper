package interfaces

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/DimKa163/keeper/internal/common"
	"github.com/DimKa163/keeper/internal/pb"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/shared/logging"
	"github.com/DimKa163/keeper/internal/server/usecase"
	"github.com/DimKa163/keeper/internal/shared"
	"github.com/beevik/guid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SyncServer struct {
	app *usecase.SyncService
	pb.UnimplementedSyncServer
}

func NewSyncServer(app *usecase.SyncService) *SyncServer {
	return &SyncServer{app: app}
}

func (ss *SyncServer) Bind(server *grpc.Server) {
	pb.RegisterSyncServer(server, ss)
}

func (ss *SyncServer) PushStream(stream pb.Sync_PushStreamServer) error {
	var v int32
	var force bool
	var err error
	ctx := stream.Context()
	v, err = common.ReadVersionFromHeader(ctx)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	force, err = common.ReadForceFromHeader(ctx)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if err = ss.app.ValidateVersion(ctx, v); err != nil && !force {
		return status.Error(codes.FailedPrecondition, err.Error())
	}
	if err = ss.app.Push(ctx, func(ctx context.Context) (*usecase.Push, error) {
		var op *pb.PushOperation
		op, err = stream.Recv()
		if err != nil {
			return nil, err
		}
		var req *usecase.Push
		switch op.GetType() {
		case pb.OperationType_Default:
			req, err = toDefault(op)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			return req, nil
		case pb.OperationType_Begin:
			req, err = toBeginFile(op)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			return req, nil
		case pb.OperationType_BinaryPart:
			req, err = toChunk(op)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			return req, nil
		case pb.OperationType_End:
			req, err = toEndFile(op)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			return req, nil
		default:
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unknown operation type: %v", op.GetType()))
		}
	}); err != nil && !errors.Is(err, io.EOF) {
		return status.Error(codes.Internal, err.Error())
	}
	var resp pb.PushResponse
	resp.SetSuccess(true)
	return stream.SendAndClose(&resp)
}

func (ss *SyncServer) Pull(ctx context.Context, request *pb.PullRequest) (*pb.PullResponse, error) {
	var response pb.PullResponse
	data, v, err := ss.app.Poll(ctx, request.GetSince())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response.SetVersion(v)
	secrets := make([]*pb.Secret, len(data))
	for i, item := range data {
		secret := toSecret1(item)
		secrets[i] = secret
	}
	response.SetSecrets(secrets)
	return &response, nil
}

func (ss *SyncServer) PullStream(in *pb.PullStreamRequest, stream pb.Sync_PullStreamServer) error {
	ctx := stream.Context()
	logger := logging.Logger(ctx)
	id, err := guid.ParseString(in.GetId())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	reader, err := ss.app.File(*id, in.GetVersion())
	defer func(file io.ReadCloser) {
		err = file.Close()
		if err != nil {
			logger.Warn("failed to close file", zap.Error(err))
		}
	}(reader)
	buffer := make([]byte, shared.MB)
	for {
		var n int
		n, err = reader.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return status.Error(codes.Internal, err.Error())
		}

		msg := toPullChunkFile(in.GetId(), buffer, n)
		if err = stream.Send(msg); err != nil {
			logger.Warn("failed to send pull chunk file", zap.Error(err))
			return status.Error(codes.Internal, err.Error())
		}
	}
	return nil
}

func toDefault(op *pb.PushOperation) (*usecase.Push, error) {
	var push usecase.Push
	secret := op.GetSecret()
	id, err := guid.ParseString(secret.GetId())
	if err != nil {
		return nil, err
	}
	push.Type = usecase.DefaultOperation
	data := &usecase.Secret{
		ID:         *id,
		ModifiedAt: secret.GetModifiedAt().AsTime(),
		Dek:        secret.GetDek(),
		Data:       secret.GetData(),
		Version:    secret.GetVersion(),
		Deleted:    secret.GetDeleted(),
	}
	switch secret.GetType() {
	case pb.SecretType_LoginPass:
		data.Type = domain.LoginPassType
	case pb.SecretType_Text:
		data.Type = domain.LoginPassType
	case pb.SecretType_BankCard:
		data.Type = domain.BankCardType
	case pb.SecretType_Binary:
		data.Type = domain.OtherType
	}
	push.Secret = data
	return &push, nil
}

//func toPullDefault(data *domain.Secret) *pb.PullOperation {
//	var op pb.PullOperation
//	var secret pb.Secret
//	secret.SetId(data.ID.String())
//	secret.SetModifiedAt(timestamppb.New(data.ModifiedAt))
//	secret.SetDek(data.Dek)
//	secret.SetData(data.Payload)
//	secret.SetDeleted(data.Deleted)
//	secret.SetIsBig(data.BigData)
//	secret.SetVersion(data.Version)
//	switch data.Type {
//	case domain.LoginPassType:
//		secret.SetType(pb.SecretType1_LoginPass)
//	case domain.TextType:
//		secret.SetType(pb.SecretType1_Text)
//	case domain.BankCardType:
//		secret.SetType(pb.SecretType1_BankCard)
//	case domain.OtherType:
//		secret.SetType(pb.SecretType1_Binary)
//	}
//	op.SetSecret(&secret)
//	op.SetType(pb.PullOperation_Default)
//	return &op
//}

//func toPullBeginFile(data *domain.Secret) *pb.PullOperation {
//	var op pb.PullOperation
//	var secret pb.Secret
//	secret.SetId(data.ID.String())
//	secret.SetModifiedAt(timestamppb.New(data.ModifiedAt))
//	secret.SetVersion(data.Version)
//	op.SetSecret(&secret)
//	op.SetType(pb.PullOperation_Default)
//	return &op
//}

func toPullChunkFile(id string, buffer []byte, n int) *pb.Chunk {
	var chunk pb.Chunk
	chunk.SetId(id)
	chunk.SetType(pb.ChunkType_FilePart)
	chunk.SetBuffer(buffer[:n])
	return &chunk
}

func toEndPullFile(id string) *pb.Chunk {
	var op pb.Chunk
	op.SetType(pb.ChunkType_EndData)
	op.SetId(id)
	return &op
}

func toEndPullStream(id string, version int32) *pb.Chunk {
	var op pb.Chunk
	op.SetType(pb.ChunkType_EndData)
	op.SetId(id)
	return &op
}

func toBeginFile(op *pb.PushOperation) (*usecase.Push, error) {
	var push usecase.Push
	secret := op.GetSecret()
	id, err := guid.ParseString(secret.GetId())
	if err != nil {
		return nil, err
	}
	push.Type = usecase.BeginOperation
	data := &usecase.Secret{
		ID:         *id,
		ModifiedAt: secret.GetModifiedAt().AsTime(),
		Version:    secret.GetVersion(),
		Type:       domain.OtherType,
	}
	push.Secret = data
	return &push, nil
}

func toChunk(op *pb.PushOperation) (*usecase.Push, error) {
	var push usecase.Push
	secret := op.GetSecret()
	id, err := guid.ParseString(secret.GetId())
	if err != nil {
		return nil, err
	}
	push.Type = usecase.ChunkOperation
	data := &usecase.Secret{
		ID: *id,
	}
	push.Secret = data
	push.Buffer = op.GetBuffer()
	return &push, nil
}

func toEndFile(op *pb.PushOperation) (*usecase.Push, error) {
	var push usecase.Push
	secret := op.GetSecret()
	id, err := guid.ParseString(secret.GetId())
	if err != nil {
		return nil, err
	}
	push.Type = usecase.EndOperation
	data := &usecase.Secret{
		ID:         *id,
		ModifiedAt: secret.GetModifiedAt().AsTime(),
		Dek:        secret.GetDek(),
		Data:       secret.GetData(),
		Version:    secret.GetVersion(),
		Type:       domain.OtherType,
		Deleted:    secret.GetDeleted(),
	}
	push.Secret = data
	return &push, nil
}

func toSecret1(data *domain.Secret) *pb.Secret {
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
		secret.SetType(pb.SecretType_LoginPass)
	case domain.TextType:
		secret.SetType(pb.SecretType_Text)
	case domain.BankCardType:
		secret.SetType(pb.SecretType_BankCard)
	case domain.OtherType:
		secret.SetType(pb.SecretType_Binary)
	}
	return &secret
}
