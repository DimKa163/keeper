package usecase

import (
	"context"
	"errors"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/infrastructure/persistence"
	"github.com/DimKa163/keeper/internal/server/shared/auth"
	"github.com/beevik/guid"
)

var (
	ErrDataConflict = errors.New("data conflict")
)

type Data struct {
	ID        guid.Guid
	Name      string
	Type      domain.StoredDataType
	Large     bool
	DekNonce  []byte
	Dek       []byte
	DataNonce []byte
	Data      []byte
	Version   int32
}

type DataService struct {
	unitOfWork domain.UnitOfWork
}

func NewDataService(uow domain.UnitOfWork) *DataService {
	return &DataService{unitOfWork: uow}
}

func (ds *DataService) Upload(ctx context.Context, model *Data) (*Data, error) {
	user, err := auth.User(ctx)
	if err != nil {
		return nil, err
	}
	data := &domain.StoredData{
		ID:        model.ID,
		Name:      model.Name,
		UserID:    user,
		Type:      model.Type,
		Large:     model.Large,
		DekNonce:  model.DekNonce,
		Dek:       model.Dek,
		DataNonce: model.DataNonce,
		Data:      model.Data,
		Version:   model.Version,
	}
	err = ds.unitOfWork.Tx(ctx, func(ctx context.Context, work domain.UnitOfWork) error {
		dataRepository := work.StoredDataRepository()
		isNew := false
		exData, err := dataRepository.Get(ctx, model.ID)
		if err != nil && !errors.Is(err, persistence.ErrResourceNotFound) {
			return err
		}
		isNew = exData == nil
		if !isNew {
			exData, err = ds.update(ctx, data, exData)
			if err != nil {
				return err
			}
			err = dataRepository.Update(ctx, exData)
			if err != nil {
				return err
			}
			return nil
		}
		err = dataRepository.Insert(ctx, data)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	dataRepository := ds.unitOfWork.StoredDataRepository()
	exData, err := dataRepository.Get(ctx, model.ID)
	if err != nil && !errors.Is(err, persistence.ErrResourceNotFound) {
		return nil, err
	}
	return &Data{
		ID:        exData.ID,
		Name:      exData.Name,
		Type:      exData.Type,
		Large:     exData.Large,
		DekNonce:  exData.DekNonce,
		Dek:       exData.Dek,
		DataNonce: exData.DataNonce,
		Data:      exData.Data,
		Version:   exData.Version,
	}, nil
}

func (ds *DataService) Download(ctx context.Context) (*DataIterator, error) {
	return NewDataIterator(ds.unitOfWork.StoredDataRepository(), 100), nil
}

func (ds *DataService) update(ctx context.Context, data, exData *domain.StoredData) (*domain.StoredData, error) {
	if exData.Compare(data) >= 0 {
		return nil, ErrDataConflict
	}
	exData.Name = data.Name
	exData.Data = data.Data
	exData.DataNonce = data.DataNonce
	exData.Dek = data.Dek
	exData.DekNonce = data.DekNonce
	exData.Large = data.Large
	exData.Version = data.Version
	return exData, nil
}

type DataIterator struct {
	repository domain.StoredDataRepository
	items      []*domain.StoredData
	limit      int
	offset     int
	index      int
}

func NewDataIterator(repository domain.StoredDataRepository, limit int) *DataIterator {
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
		ID:        item.ID,
		Name:      item.Name,
		Type:      item.Type,
		Large:     item.Large,
		DekNonce:  item.DekNonce,
		Dek:       item.Dek,
		DataNonce: item.DataNonce,
		Data:      item.Data,
		Version:   item.Version,
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
