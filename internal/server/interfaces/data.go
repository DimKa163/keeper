package interfaces

import (
	"context"
	"errors"
	"github.com/DimKa163/keeper/internal/pb"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/shared/auth"
	"github.com/DimKa163/keeper/internal/server/shared/logging"
	"github.com/DimKa163/keeper/internal/shared"
	"github.com/beevik/guid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"os"
)

type DataServer struct {
	app domain.DataService
	pb.UnimplementedSyncDataServer
}

func NewDataServer(app domain.DataService) *DataServer {
	return &DataServer{app: app}
}

func (ds *DataServer) Bind(server *grpc.Server) {
	pb.RegisterSyncDataServer(server, ds)
}
func (ds *DataServer) PushUnary(ctx context.Context, request *pb.PushRequest) (*pb.PushResponse, error) {
	var response pb.PushResponse
	if err := validateBatchUploadRequest(request); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dataSlice := make([]*domain.Data, len(request.GetData()))
	for i, d := range request.GetData() {
		data, err := toIn(ctx, d)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		dataSlice[i] = data
	}

	err := ds.app.Push(ctx, dataSlice)
	if err != nil {
		if errors.Is(err, domain.ErrDataConflict) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	response.SetSuccess(true)
	return &response, nil
}

func (ds *DataServer) PushStream(stream pb.SyncData_PushStreamServer) error {
	for {
		ctx := stream.Context()
		logger := logging.Logger(ctx)
		var resp pb.PushResponse
		logger.Debug("start push stream request")
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Debug("finish push stream request")
				return stream.SendAndClose(&resp)
			}
			logger.Error("failed to receive push stream request", zap.Error(err))
			return status.Error(codes.Internal, err.Error())
		}
		logger = logger.With(zap.String("type", req.GetType().String()))
		logger.Debug("receive data")
		ctx = logging.SetLogger(ctx, logger)
		switch req.GetType() {
		case pb.RequestType_Default:
			data, err := toIn(ctx, req.GetData())
			if err != nil {
				logger.Error("validation failed", zap.Error(err))
				return status.Error(codes.InvalidArgument, err.Error())
			}
			if err := ds.app.PushUnary(ctx, data); err != nil {
				logger.Error("failed to push data", zap.Error(err))
				return status.Error(codes.Internal, err.Error())
			}
		case pb.RequestType_StartData:
			data, err := toIn(ctx, req.GetData())
			if err != nil {
				logger.Error("validation failed", zap.Error(err))
				return status.Error(codes.InvalidArgument, err.Error())
			}
			if err := ds.app.PushMetadata(ctx, data); err != nil {
				logger.Error("failed pushing metadata", zap.Error(err))
				return status.Error(codes.Internal, err.Error())
			}
		case pb.RequestType_FilePart:
			id, err := guid.ParseString(req.GetData().GetId())
			if err != nil {
				logger.Error("validation failed", zap.Error(err))
				return status.Error(codes.InvalidArgument, err.Error())
			}
			if err := ds.app.PushData(ctx, *id, req.GetChunk()); err != nil {
				logger.Error("failed pushing data", zap.Error(err))
				return status.Error(codes.Internal, err.Error())
			}
		case pb.RequestType_EndData:
			id, err := guid.ParseString(req.GetData().GetId())
			if err != nil {
				logger.Error("validation failed", zap.Error(err))
				return status.Error(codes.InvalidArgument, err.Error())
			}
			if err := ds.app.Finish(ctx, *id); err != nil {
				logger.Error("failed finish data", zap.Error(err))
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
}

func (ds *DataServer) PollUnary(ctx context.Context, request *pb.PollRequest) (*pb.PollResponse, error) {
	var response pb.PollResponse
	data, version, err := ds.app.Poll(ctx, request.GetSince())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbData := make([]*pb.Data, len(data))
	for i, d := range data {
		pbData[i] = toOut(d)
	}
	response.SetData(pbData)
	response.SetServerVersion(version)
	return &response, nil
}

func (ds *DataServer) PollStream(in *pb.PollRequest, stream pb.SyncData_PollStreamServer) error {
	data, version, err := ds.app.Poll(stream.Context(), in.GetSince())
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	for _, d := range data {
		var poll pb.Poll
		poll.SetServerVersion(version)
		if d.BigData {
			out := toOut(d)
			poll.SetType(pb.RequestType_StartData)
			poll.SetData(out)
			if err = stream.Send(&poll); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			file, err := d.File()
			if err != nil {
				if !os.IsNotExist(err) {
					return status.Error(codes.Internal, err.Error())
				}
				poll = pb.Poll{}
				poll.SetServerVersion(version)
				poll.SetType(pb.RequestType_ErrData)
				poll.SetData(out)
				continue
			}
			defer file.Close()
			buffer := make([]byte, shared.MB)
			for {
				poll = pb.Poll{}
				poll.SetServerVersion(version)
				poll.SetType(pb.RequestType_FilePart)
				poll.SetData(out)

				n, err := file.Read(buffer)
				if err != nil {
					if err == io.EOF {
						break
					}
					return status.Error(codes.Internal, err.Error())
				}
				if n == 0 {
					break
				}
				poll.SetChunk(buffer[:n])
				if err = stream.Send(&poll); err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			}
			poll = pb.Poll{}
			poll.SetServerVersion(version)
			poll.SetType(pb.RequestType_EndData)
			poll.SetData(out)

		} else {
			poll.SetData(toOut(d))
			poll.SetType(pb.RequestType_Default)
			if err = stream.Send(&poll); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
	return nil
}

func validateBatchUploadRequest(req *pb.PushRequest) error {
	err := make([]error, 0)
	data := req.GetData()
	if len(data) == 0 {
		err = append(err, errors.New("data empty"))
	} else {
		for _, d := range data {
			if errs := validateData(d); errs != nil {
				err = append(err, errs)
			}
		}
	}
	if len(err) > 0 {
		return errors.Join(err...)
	}
	return nil
}

func validateData(data *pb.Data) error {
	err := make([]error, 0)
	if !data.HasData() {
		err = append(err, errors.New("stored data is required"))
	}
	if !data.HasDataNonce() {
		err = append(err, errors.New("stored data nonce is required"))
	}
	if !data.HasDek() {
		err = append(err, errors.New("dek is required"))
	}
	if !data.HasDekNonce() {
		err = append(err, errors.New("dek nonce is required"))
	}
	if !data.HasType() {
		err = append(err, errors.New("invalid type"))
	}
	if !data.HasVersion() {
		err = append(err, errors.New("version is required"))
	}
	if len(err) > 0 {
		return errors.Join(err...)
	}
	return nil
}

func toIn(ctx context.Context, data *pb.Data) (*domain.Data, error) {
	dt := domain.Data{}
	id, err := guid.ParseString(data.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := auth.User(ctx)
	if err != nil {
		return nil, err
	}
	dt.ID = *id
	dt.BigData = data.GetLarge()
	dt.Payload = data.GetData()
	dt.PayloadNonce = data.GetDataNonce()
	dt.UserID = user
	dt.Dek = data.GetDek()
	dt.DekNonce = data.GetDekNonce()
	dt.FileDataNonce = data.GetFileDataNonce()
	dt.Version = data.GetVersion()
	switch data.GetType() {
	case pb.Data_LoginPass:
		dt.Type = domain.LoginPassType
	case pb.Data_Text:
		dt.Type = domain.TextType
	case pb.Data_BankCard:
		dt.Type = domain.BankCardType
	case pb.Data_Other:
		dt.Type = domain.OtherType
	}
	return &dt, nil
}

func toOut(op *domain.Data) *pb.Data {
	data := &pb.Data{}
	data.SetId(op.ID.String())
	data.SetVersion(op.Version)
	data.SetData(op.Payload)
	data.SetDataNonce(op.PayloadNonce)
	data.SetDek(op.Dek)
	data.SetDekNonce(op.DekNonce)
	data.SetFileDataNonce(op.FileDataNonce)
	data.SetLarge(op.BigData)
	switch op.Type {
	case domain.LoginPassType:
		data.SetType(pb.Data_LoginPass)
	case domain.TextType:
		data.SetType(pb.Data_Text)
	case domain.BankCardType:
		data.SetType(pb.Data_BankCard)
	case domain.OtherType:
		data.SetType(pb.Data_Other)
	}
	return data
}
