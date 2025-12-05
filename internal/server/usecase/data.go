package usecase

import (
	"context"
	"errors"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/infrastructure/persistence"
	"github.com/DimKa163/keeper/internal/server/shared/auth"
	"github.com/DimKa163/keeper/internal/server/shared/logging"
	"github.com/DimKa163/keeper/internal/shared"
	"github.com/beevik/guid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"io"
	"os"
	"reflect"
	"time"
)

var syncTypeName = reflect.TypeOf(domain.Data{}).Name()

var (
	ErrDataConflict = errors.New("data conflict")
)

type (
	Secret struct {
		ID         guid.Guid `json:"id"`
		ModifiedAt time.Time
		UserID     guid.Guid `json:"user_id"`
		Type       domain.DataType
		BigData    bool   `json:"big_data"`
		Dek        []byte `json:"dek"`
		Payload    []byte `json:"payload"`
		Deleted    bool   `json:"deleted"`
		Version    int32  `json:"version"`
	}
	PushUnaryRequest struct {
		Secrets []*Secret `json:"secrets"`
		Force   bool
	}
	PushMetadataRequest struct {
		ID      guid.Guid `json:"id"`
		UserID  guid.Guid `json:"user_id"`
		Type    domain.DataType
		BigData bool `json:"big_data"`
		Deleted bool `json:"deleted"`
	}
	PushChunkRequest struct {
		ID         guid.Guid `json:"id"`
		ModifiedAt time.Time `json:"modified_at"`
		Buffer     []byte    `json:"buffer"`
	}
	PushFileCloseRequest struct {
		ID         guid.Guid `json:"id"`
		ModifiedAt time.Time `json:"modified_at"`
		Dek        []byte    `json:"dek"`
		Payload    []byte    `json:"payload"`
	}
)

type DataService struct {
	unitOfWork   domain.UnitOfWork
	fileProvider *shared.FileProvider
}

func NewDataService(uow domain.UnitOfWork, fileProvider *shared.FileProvider) *DataService {
	return &DataService{unitOfWork: uow, fileProvider: fileProvider}
}

func (ds *DataService) Push(ctx context.Context, request *PushUnaryRequest) error {
	if len(request.Secrets) == 0 {
		return errors.New("empty request")
	}
	logger := logging.Logger(ctx)
	logger.Info("pushing request")
	if err := ds.unitOfWork.Tx(ctx, func(ctx context.Context, work domain.UnitOfWork) error {
		logger := logging.Logger(ctx)
		userId, err := auth.User(ctx)
		if err != nil {
			return err
		}
		stateRepository := work.SyncStateRepository()
		state, err := stateRepository.Get(ctx, syncTypeName, userId)
		if err != nil {
			return err
		}
		logger = logger.With(zap.Int32("version", state.Value))
		ctx = logging.SetLogger(ctx, logger)
		state.Value += 1
		for _, secret := range request.Secrets {
			if secret.BigData {
				if err = ds.applyFile(ctx, work, state, secret, request.Force); err != nil {
					return err
				}
			} else {
				if err = ds.apply(ctx, work, state, secret, request.Force); err != nil {
					return err
				}
			}
		}

		if err = stateRepository.Update(ctx, state); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (ds *DataService) applyFile(ctx context.Context, uow domain.UnitOfWork, state *domain.SyncState, secret *Secret, force bool) error {
	logger := logging.Logger(ctx)
	dataRepository := uow.DataRepository()
	data, err := dataRepository.Get(ctx, secret.ID)
	if err != nil && !errors.Is(err, persistence.ErrResourceNotFound) {
		return err
	}
	if data != nil {
		if secret.Version <= data.Version && !force {
			logger.Warn("conflict! server win")
			return ErrDataConflict
		}
		data.Deleted = secret.Deleted
		if data.Deleted {
			if err = ds.fileProvider.Remove(data.ID.String(), data.Version); err != nil {
				return err
			}
			data.ModifiedAt = secret.ModifiedAt
			data.Version = state.Value
		}
		return dataRepository.Update(ctx, data)
	}
	data = &domain.Data{
		ID:         secret.ID,
		ModifiedAt: secret.ModifiedAt,
		UserID:     secret.UserID,
		Type:       secret.Type,
		BigData:    secret.BigData,
	}
	return dataRepository.Insert(ctx, data)
}

func (ds *DataService) apply(ctx context.Context, uow domain.UnitOfWork, state *domain.SyncState, secret *Secret, force bool) error {

	dataRepository := uow.DataRepository()
	data, err := dataRepository.Get(ctx, secret.ID)
	if err != nil && !errors.Is(err, persistence.ErrResourceNotFound) {
		return err
	}
	if data != nil {
		// no change
		if data.Version == secret.Version && data.ModifiedAt.Equal(secret.ModifiedAt) {
			return nil
		}
		logger := logging.Logger(ctx)
		if secret.Version <= data.Version && !force {
			logger.Warn("conflict! server win")
			return ErrDataConflict
		}
		data.ModifiedAt = secret.ModifiedAt
		data.Dek = secret.Dek
		data.Payload = secret.Payload
		data.Version = state.Value
		data.Deleted = secret.Deleted
		return dataRepository.Update(ctx, data)
	}
	data = &domain.Data{
		ID:         secret.ID,
		ModifiedAt: secret.ModifiedAt,
		UserID:     secret.UserID,
		Type:       secret.Type,
		BigData:    secret.BigData,
		Dek:        secret.Dek,
		Payload:    secret.Payload,
		Deleted:    secret.Deleted,
		Version:    state.Value,
	}
	return dataRepository.Insert(ctx, data)
}

func (ds *DataService) HandleChunk(ctx context.Context, req *PushChunkRequest) error {
	logger := logging.Logger(ctx)
	logger.Info("pushing part data")
	userId, err := auth.User(ctx)
	if err != nil {
		return err
	}
	repository := ds.unitOfWork.DataRepository()
	data, err := repository.Get(ctx, req.ID)
	if err != nil {
		return err
	}
	stateRepository := ds.unitOfWork.SyncStateRepository()
	state, err := stateRepository.Get(ctx, syncTypeName, userId)
	if err != nil {
		return err
	}
	f, err := ds.fileProvider.OpenWrite(data.ID.String(), state.Value)
	if err != nil {
		return err
	}
	defer func(f io.WriteCloser) {
		err = f.Close()
		if err != nil {
			logger.Warn("failed to close file", zap.Error(err))
		}
	}(f)
	_, err = f.Write(req.Buffer)
	if err != nil {
		return err
	}
	return nil
}

func (ds *DataService) Commit(ctx context.Context, req *PushFileCloseRequest) error {
	if err := ds.unitOfWork.Tx(ctx, func(ctx context.Context, work domain.UnitOfWork) error {
		logger := logging.Logger(ctx)
		logger.Info("pushing final data")
		dataRepository := work.DataRepository()
		data, err := dataRepository.Get(ctx, req.ID)
		if err != nil {
			return err
		}
		oldVersion := data.Version
		userId, err := auth.User(ctx)
		if err != nil {
			return err
		}
		stateRepository := work.SyncStateRepository()
		state, err := stateRepository.Get(ctx, syncTypeName, userId)
		if err != nil {
			return err
		}
		data.Dek = req.Dek
		data.Payload = req.Payload
		data.Version = state.Value
		data.ModifiedAt = req.ModifiedAt
		if err = dataRepository.Update(ctx, data); err != nil {
			return err
		}
		// удаляем старую версию
		if err = ds.fileProvider.Remove(data.ID.String(), oldVersion); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (ds *DataService) OpenFile(id string, version int32) (io.ReadCloser, error) {
	return ds.fileProvider.OpenRead(id, version)
}
func (ds *DataService) Poll(ctx context.Context, since int32) ([]*domain.Data, int32, error) {
	user, err := auth.User(ctx)
	if err != nil {
		return nil, -1, err
	}
	rep := ds.unitOfWork.DataRepository()
	data, err := rep.GetAll(ctx, user, since)
	if err != nil {
		return nil, -1, err
	}
	userId, err := auth.User(ctx)
	if err != nil {
		return nil, -1, err
	}
	syncRepository := ds.unitOfWork.SyncStateRepository()
	state, err := syncRepository.Get(ctx, syncTypeName, userId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, -1, err
		}
		state = &domain.SyncState{
			ID:    syncTypeName,
			Value: 0,
		}
	}
	return data, state.Value, nil
}
