package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/infrastructure/persistence"
	"github.com/DimKa163/keeper/internal/server/shared/auth"
	"github.com/DimKa163/keeper/internal/server/shared/logging"
	"github.com/DimKa163/keeper/internal/shared"
	"github.com/beevik/guid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"reflect"
)

var syncTypeName = reflect.TypeOf(domain.Data{}).Name()

var (
	ErrDataConflict = errors.New("data conflict")
)

type DataService struct {
	unitOfWork   domain.UnitOfWork
	fileProvider *shared.FileProvider
}

func NewDataService(uow domain.UnitOfWork, fileProvider *shared.FileProvider) *DataService {
	return &DataService{unitOfWork: uow, fileProvider: fileProvider}
}

func (ds *DataService) PushUnary(ctx context.Context, data *domain.Data) error {
	logger := logging.Logger(ctx)
	logger.Info("pushing data")
	if err := ds.unitOfWork.Tx(ctx, func(ctx context.Context, work domain.UnitOfWork) error {
		logger := logging.Logger(ctx)
		userId, err := auth.User(ctx)
		if err != nil {
			return err
		}
		stateRepository := work.SyncStateRepository()
		state, err := stateRepository.Get(ctx, syncTypeName, userId)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				logger.Error("failed to get state", zap.Error(err))
				return err
			}
			state = &domain.SyncState{
				ID:    syncTypeName,
				Value: 0,
			}
		}
		logger = logger.With(zap.Int32("version", state.Value))
		ctx = logging.SetLogger(ctx, logger)
		dataRepository := work.DataRepository()
		if err = ds.process(ctx, state, dataRepository, data); err != nil {
			return err
		}
		if err = stateRepository.Save(ctx, state); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (ds *DataService) PushMetadata(ctx context.Context, model *domain.Data) error {
	logger := logging.Logger(ctx)
	repository := ds.unitOfWork.DataRepository()
	data, err := repository.Get(ctx, model.ID)
	if err != nil && !errors.Is(err, persistence.ErrResourceNotFound) {
		return err
	}
	if data == nil {
		logger.Info("data does not exist")
		model.Path = fmt.Sprintf("%s_%d", model.ID.String(), model.Version)
		if err = repository.Insert(ctx, model); err != nil {
			return err
		}
		return nil
	}

	if err = ds.fileProvider.Remove(fmt.Sprintf("%s_%d", data.Path, data.Version)); err != nil {
		return err
	}

	data.Update(
		model.BigData,
		model.DekNonce,
		model.Dek,
		model.PayloadNonce,
		model.Payload,
		model.FileDataNonce,
		model.Deleted,
		data.Version,
	)
	if err = repository.Update(ctx, data); err != nil {
		return err
	}
	return nil
}

func (ds *DataService) PushData(ctx context.Context, id guid.Guid, payload []byte) error {
	logger := logging.Logger(ctx)
	logger.Info("pushing part data")
	userId, err := auth.User(ctx)
	if err != nil {
		return err
	}
	repository := ds.unitOfWork.DataRepository()
	data, err := repository.Get(ctx, id)
	if err != nil {
		return err
	}
	stateRepository := ds.unitOfWork.SyncStateRepository()
	state, err := stateRepository.Get(ctx, syncTypeName, userId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		state = &domain.SyncState{
			ID:     syncTypeName,
			UserID: userId,
			Value:  0,
		}
	}
	state.Value += 1
	f, err := ds.fileProvider.OpenWrite(data.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(payload)
	if err != nil {
		return err
	}
	return nil
}

func (ds *DataService) Finish(ctx context.Context, id guid.Guid) error {
	if err := ds.unitOfWork.Tx(ctx, func(ctx context.Context, work domain.UnitOfWork) error {
		logger := logging.Logger(ctx)
		logger.Info("pushing final data")
		userId, err := auth.User(ctx)
		if err != nil {
			return err
		}
		stateRepository := work.SyncStateRepository()
		state, err := stateRepository.Get(ctx, syncTypeName, userId)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			state = &domain.SyncState{
				ID:     syncTypeName,
				UserID: userId,
				Value:  0,
			}
		}
		state.Value += 1
		dataRepository := work.DataRepository()
		data, err := dataRepository.Get(ctx, id)
		if err != nil {
			return err
		}
		data.Version = state.Value
		if err = dataRepository.Update(ctx, data); err != nil {
			return err
		}
		if err = stateRepository.Save(ctx, state); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (ds *DataService) Push(ctx context.Context, dataList []*domain.Data) error {
	//dataList = keepLast(dataList)
	if err := ds.unitOfWork.Tx(ctx, func(ctx context.Context, work domain.UnitOfWork) error {
		userId, err := auth.User(ctx)
		if err != nil {
			return err
		}
		stateRepository := work.SyncStateRepository()
		state, err := stateRepository.Get(ctx, syncTypeName, userId)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			state = &domain.SyncState{
				ID:    syncTypeName,
				Value: 0,
			}
		}
		dataRepository := work.DataRepository()
		for _, data := range dataList {
			if err = ds.process(ctx, state, dataRepository, data); err != nil {
				return err
			}
		}
		if err = stateRepository.Save(ctx, state); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
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

func (ds *DataService) process(ctx context.Context, state *domain.SyncState, repository domain.DataRepository, model *domain.Data) error {
	logger := logging.Logger(ctx)
	logger.Info("processing data")
	data, err := repository.Get(ctx, model.ID)
	if err != nil && !errors.Is(err, persistence.ErrResourceNotFound) {
		return err
	}
	state.Value += 1
	if data != nil {
		logger.Info(fmt.Sprintf("updating data with id %d", data.ID))
		if model.Version < data.Version {
			logger.Warn("conflict! server win")
			return nil
		}
		data.Update(
			model.BigData,
			model.DekNonce,
			model.Dek,
			model.PayloadNonce,
			model.Payload,
			model.FileDataNonce,
			model.Deleted,
			state.Value,
		)
		err = repository.Update(ctx, data)
		if err != nil {
			return err
		}
		return nil
	}
	logger.Info(fmt.Sprintf("creating data with id %s", model.ID.String()))
	data = model
	data.Version = state.Value
	if err = repository.Insert(ctx, data); err != nil {
		return err
	}
	return nil
}
