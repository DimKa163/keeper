package interfaces

import (
	"context"
	"errors"
	"github.com/DimKa163/keeper/internal/server/shared/auth"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/interfaces/pb"
	"github.com/DimKa163/keeper/internal/server/usecase"
	"github.com/beevik/guid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DataServer struct {
	app *usecase.DataService
	pb.UnimplementedStoredDataServer
}

func NewDataServer(app *usecase.DataService) *DataServer {
	return &DataServer{app: app}
}

func (ds *DataServer) Bind(server *grpc.Server) {
	pb.RegisterStoredDataServer(server, ds)
}
func (ds *DataServer) Push(ctx context.Context, request *pb.PushRequest) (*pb.PushResponse, error) {
	var response pb.PushResponse
	if err := validateBatchUploadRequest(request); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dataSlice := make([]*domain.Operation, len(request.GetRequests()))
	for i, d := range request.GetRequests() {
		data, err := toIn(ctx, d)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		dataSlice[i] = data
	}

	result, err := ds.app.Push(ctx, dataSlice)
	if err != nil {
		if errors.Is(err, domain.ErrDataConflict) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := make([]*pb.OperationResponse, len(result))
	for i, d := range result {
		resp[i] = toOut(d)
	}
	response.SetData(resp)
	return &response, nil
}

func (ds *DataServer) Poll(_ *pb.PollRequest, srv pb.StoredData_PollServer) error {
	iterator := ds.app.GetIterator()
	next, err := iterator.MoveNext(srv.Context())
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	for next {
		it := iterator.Current()
		resp := toOut(it)
		if err := srv.Send(resp.GetData()); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		next, err = iterator.MoveNext(srv.Context())
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
	return nil
}

func validateBatchUploadRequest(req *pb.PushRequest) error {
	err := make([]error, 0)
	data := req.GetRequests()
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

func validateData(operation *pb.Operation) error {
	err := make([]error, 0)
	if !operation.HasData() {
		err = append(err, errors.New("stored data is required"))
	}
	data := operation.GetData()
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

func toIn(ctx context.Context, op *pb.Operation) (*domain.Operation, error) {
	data := op.GetData()
	oper := domain.Operation{}
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
	dt.Name = data.GetName()
	dt.Large = data.GetLarge()
	dt.Payload = data.GetData()
	dt.PayloadNonce = data.GetDataNonce()
	dt.UserID = user
	dt.Dek = data.GetDek()
	dt.DekNonce = data.GetDekNonce()
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
	oper.Data = &dt
	return &oper, nil
}

func toOut(op *domain.Data) *pb.OperationResponse {
	resp := &pb.OperationResponse{}
	data := &pb.Data{}
	data.SetId(op.ID.String())
	data.SetName(op.Name)
	data.SetVersion(op.Version)
	data.SetData(op.Payload)
	data.SetDataNonce(op.PayloadNonce)
	data.SetDek(op.Dek)
	data.SetDekNonce(op.DekNonce)
	data.SetLarge(op.Large)
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
	return resp
}
