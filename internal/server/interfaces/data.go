package interfaces

import (
	"context"
	"errors"

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

func (ds *DataServer) Upload(ctx context.Context, request *pb.UploadRequest) (*pb.UploadResponse, error) {
	var response pb.UploadResponse
	if err := validateUpload(request); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	dt := request.GetData()
	data, err := toIn(dt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	sD, err := ds.app.Upload(ctx, data)
	if err != nil {
		if errors.Is(err, domain.ErrDataConflict) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	response.SetData(toOut(sD))
	return &response, nil
}

func (ds *DataServer) BatchUpload(ctx context.Context, request *pb.BatchUploadRequest) (*pb.BatchUploadResponse, error) {
	var response pb.BatchUploadResponse
	if err := validateBatchUploadRequest(request); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dataSlice := make([]*usecase.Data, len(request.GetData()))
	for i, d := range request.GetData() {
		data, err := toIn(d)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		dataSlice[i] = data
	}

	result, err := ds.app.BatchUpload(ctx, dataSlice)
	if err != nil {
		if errors.Is(err, domain.ErrDataConflict) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := make([]*pb.Data, len(result))
	for i, d := range result {
		resp[i] = toOut(d)
	}
	response.SetData(resp)
	return &response, nil
}

func validateUpload(req *pb.UploadRequest) error {
	err := make([]error, 0)
	if !req.HasData() {
		err = append(err, errors.New("data is required"))
	} else {
		data := req.GetData()
		if errs := validateData(data); errs != nil {
			err = append(err, errs)
		}
	}
	if len(err) > 0 {
		return errors.Join(err...)
	}
	return nil
}

func validateBatchUploadRequest(req *pb.BatchUploadRequest) error {
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

func toIn(data *pb.Data) (*usecase.Data, error) {
	dt := usecase.Data{}
	id, err := guid.ParseString(data.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	dt.ID = *id
	dt.Name = data.GetName()
	dt.Large = data.GetLarge()
	dt.Payload = data.GetData()
	dt.PayloadNonce = data.GetDataNonce()
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
	return &dt, nil
}

func toOut(data *usecase.Data) *pb.Data {
	resp := &pb.Data{}
	resp.SetId(data.ID.String())
	resp.SetName(data.Name)
	resp.SetVersion(data.Version)
	resp.SetData(data.Payload)
	resp.SetDataNonce(data.PayloadNonce)
	resp.SetDek(data.Dek)
	resp.SetDekNonce(data.DekNonce)
	resp.SetLarge(data.Large)
	switch data.Type {
	case domain.LoginPassType:
		resp.SetType(pb.Data_LoginPass)
	case domain.TextType:
		resp.SetType(pb.Data_Text)
	case domain.BankCardType:
		resp.SetType(pb.Data_BankCard)
	case domain.OtherType:
		resp.SetType(pb.Data_Other)
	}
	return resp
}
