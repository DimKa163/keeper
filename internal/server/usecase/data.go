package usecase

import (
	"context"
	"errors"

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

func (ds *DataService) GetIterator() domain.DataIterator {
	return NewDataIterator(ds.unitOfWork.DataRepository(), 100)
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
		if err := repository.Delete(ctx, model.Data.ID); err != nil {
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

type DataIterator struct {
	repository domain.DataRepository
	items      []*domain.Data
	limit      int
	offset     int
	index      int
	current    *domain.Data
}

func NewDataIterator(repository domain.DataRepository, limit int) *DataIterator {
	return &DataIterator{
		repository: repository,
		limit:      limit,
		offset:     0,
		current:    nil,
		index:      0,
	}
}

func (di *DataIterator) MoveNext(ctx context.Context) (bool, error) {
	if di.items == nil || len(di.items) <= di.index {
		di.current = nil
		if err := di.load(ctx); err != nil {
			return false, err
		}
	}
	if len(di.items) == 0 {
		di.current = nil
		return false, nil
	}
	di.current = di.items[di.index]
	di.index += 1
	return true, nil
}

func (di *DataIterator) Current() *domain.Data {
	return di.current
}

func (di *DataIterator) load(ctx context.Context) error {
	var err error
	user, err := auth.User(ctx)
	if err != nil {
		return err
	}
	all, err := di.repository.GetAll(ctx, user, di.limit, di.offset)
	di.items = make([]*domain.Data, len(all))
	for i, item := range all {
		di.items[i] = item
	}
	if err != nil {
		return err
	}
	di.offset = len(di.items)
	di.index = 0
	return nil
}
