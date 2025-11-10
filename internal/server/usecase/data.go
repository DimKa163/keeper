package usecase

import (
	"context"
	"errors"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/infrastructure/persistence"
	"github.com/DimKa163/keeper/internal/server/shared/auth"
	"github.com/beevik/guid"
)

type Data struct {
	ID           guid.Guid
	Name         string
	Type         domain.DataType
	Large        bool
	DekNonce     []byte
	Dek          []byte
	PayloadNonce []byte
	Payload      []byte
	Version      int32
}

type DataService struct {
	unitOfWork domain.UnitOfWork
}

func NewDataService(uow domain.UnitOfWork) *DataService {
	return &DataService{unitOfWork: uow}
}

func (ds *DataService) Upload(ctx context.Context, model *Data) (*Data, error) {
	if err := ds.unitOfWork.Tx(ctx, func(ctx context.Context, work domain.UnitOfWork) error {
		dataRepository := work.DataRepository()
		if err := ds.process(ctx, dataRepository, model); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	dataRepository := ds.unitOfWork.DataRepository()
	exData, err := dataRepository.Get(ctx, model.ID)
	if err != nil && !errors.Is(err, persistence.ErrResourceNotFound) {
		return nil, err
	}
	return &Data{
		ID:           exData.ID,
		Name:         exData.Name,
		Type:         exData.Type,
		Large:        exData.Large,
		DekNonce:     exData.DekNonce,
		Dek:          exData.Dek,
		PayloadNonce: exData.PayloadNonce,
		Payload:      exData.Payload,
		Version:      exData.Version,
	}, nil
}

func (ds *DataService) BatchUpload(ctx context.Context, dataSlice []*Data) ([]*Data, error) {
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
	resp := make([]*Data, len(dataSlice))
	for i, data := range dataSlice {
		exData, err := dataRepository.Get(ctx, data.ID)
		if err != nil {
			return nil, err
		}
		resp[i] = &Data{
			ID:           exData.ID,
			Name:         exData.Name,
			Type:         exData.Type,
			Large:        exData.Large,
			DekNonce:     exData.DekNonce,
			Dek:          exData.Dek,
			PayloadNonce: exData.PayloadNonce,
			Payload:      exData.Payload,
			Version:      exData.Version,
		}
	}
	return resp, nil
}

func (ds *DataService) Download(ctx context.Context) (*DataIterator, error) {
	return NewDataIterator(ds.unitOfWork.DataRepository(), 100), nil
}

func (ds *DataService) process(ctx context.Context, repository domain.DataRepository, model *Data) error {
	isNew := false
	exData, err := repository.Get(ctx, model.ID)
	if err != nil && !errors.Is(err, persistence.ErrResourceNotFound) {
		return err
	}
	isNew = exData == nil
	if !isNew {
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
	}
	user, err := auth.User(ctx)
	if err != nil {
		return err
	}
	data := &domain.Data{
		ID:           model.ID,
		Name:         model.Name,
		UserID:       user,
		Type:         model.Type,
		Large:        model.Large,
		DekNonce:     model.DekNonce,
		Dek:          model.Dek,
		PayloadNonce: model.PayloadNonce,
		Payload:      model.Payload,
		Version:      model.Version,
	}
	err = repository.Insert(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func keepLast(dataSlice []*Data) []*Data {
	seen := make(map[guid.Guid]bool)
	result := make([]*Data, 0, len(dataSlice))
	for i := len(dataSlice) - 1; i >= 0; i-- {
		v := dataSlice[i]
		if seen[v.ID] {
			continue
		}
		result = append(result, v)
		seen[v.ID] = true
	}
	return result
}

type DataIterator struct {
	repository domain.DataRepository
	items      []*domain.Data
	limit      int
	offset     int
	index      int
}

func NewDataIterator(repository domain.DataRepository, limit int) *DataIterator {
	return &DataIterator{repository: repository, limit: limit}
}

func (it *DataIterator) Next(ctx context.Context) (*Data, error) {
	if it.items == nil || len(it.items) <= it.index {
		if err := it.load(ctx); err != nil {
			return nil, err
		}
	}
	item := it.items[it.index]
	it.index++
	return &Data{
		ID:           item.ID,
		Name:         item.Name,
		Type:         item.Type,
		Large:        item.Large,
		DekNonce:     item.DekNonce,
		Dek:          item.Dek,
		PayloadNonce: item.PayloadNonce,
		Payload:      item.Payload,
		Version:      item.Version,
	}, nil
}

func (it *DataIterator) load(ctx context.Context) error {
	var err error
	user, err := auth.User(ctx)
	if err != nil {
		return err
	}
	it.items, err = it.repository.GetAll(ctx, user, it.limit, it.offset)
	if err != nil {
		return err
	}
	it.offset = len(it.items)
	it.index = 0
	return nil
}
