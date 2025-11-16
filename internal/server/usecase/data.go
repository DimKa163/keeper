package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/infrastructure/persistence"
	"github.com/DimKa163/keeper/internal/server/shared/auth"
	"github.com/beevik/guid"
)

type DataService struct {
	unitOfWork domain.UnitOfWork
}

func NewDataService(uow domain.UnitOfWork) *DataService {
	return &DataService{unitOfWork: uow}
}

func (ds *DataService) Push(ctx context.Context, dataSlice []*domain.Operation) ([]*domain.Data, error) {
	dataSlice = keepLast(dataSlice)
	if err := ds.unitOfWork.Tx(ctx, func(ctx context.Context, work domain.UnitOfWork) error {
		dataRepository := work.DataRepository()
		for _, data := range dataSlice {
			if err := ds.process(ctx, dataRepository, data); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	dataRepository := ds.unitOfWork.DataRepository()
	resp := make([]*domain.Data, len(dataSlice))
	for i, data := range dataSlice {
		exData, err := dataRepository.Get(ctx, data.ID)
		if err != nil {
			return nil, err
		}
		resp[i] = exData
	}
	return resp, nil
}

func (ds *DataService) Poll(ctx context.Context, since time.Time) ([]*domain.Data, error) {
	user, err := auth.User(ctx)
	if err != nil {
		return nil, err
	}
	rep := ds.unitOfWork.DataRepository()
	data, err := rep.GetAll(ctx, user, since)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (ds *DataService) process(ctx context.Context, repository domain.DataRepository, model *domain.Operation) error {
	switch model.OperationType {
	case domain.InsertOperation:
		data := model.Data
		err := repository.Insert(ctx, data)
		if err != nil {
			return err
		}
	case domain.UpdateOperation:
		exData, err := repository.Get(ctx, model.Data.ID)
		if err != nil && !errors.Is(err, persistence.ErrResourceNotFound) {
			return err
		}
		if err = exData.Update(
			model.Name,
			model.Large,
			model.DekNonce,
			model.Dek,
			model.PayloadNonce,
			model.Payload,
			model.Version,
		); err != nil {
			return err
		}
		exData.UpVersion()
		err = repository.Update(ctx, exData)
		if err != nil {
			return err
		}
		return nil
	case domain.DeleteOperation:
		model.Deleted = true
		model.UpVersion()
		if err := repository.Delete(ctx, model.Data); err != nil {
			return err
		}
	}
	return nil
}

func keepLast(dataSlice []*domain.Operation) []*domain.Operation {
	seen := make(map[guid.Guid]bool)
	result := make([]*domain.Operation, 0, len(dataSlice))
	for i := len(dataSlice) - 1; i >= 0; i-- {
		v := dataSlice[i]
		if seen[v.Data.ID] {
			continue
		}
		result = append(result, v)
		seen[v.Data.ID] = true
	}
	return result
}
