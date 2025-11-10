package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/DimKa163/keeper/internal/mocks"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/infrastructure/persistence"
	"github.com/DimKa163/keeper/internal/server/shared/auth"
	"github.com/beevik/guid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// TestDataService_Upload_Should_CreateDataSuccessfully просто создаем новые данные
func TestDataService_Upload_Should_CreateDataSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	ctx = auth.SetUser(ctx, *guid.New())
	id, data, local := generateData("some pass", domain.LoginPassType)
	uow := mocks.NewMockUnitOfWork(ctrl)
	mockRep := mocks.NewMockDataRepository(ctrl)
	mockRep.EXPECT().Get(ctx, id).Return(nil, persistence.ErrResourceNotFound)
	mockRep.EXPECT().Get(ctx, id).Return(local, nil)
	mockRep.EXPECT().Insert(ctx, gomock.Any()).Return(nil)
	uow.EXPECT().DataRepository().Return(mockRep).Times(2)
	uow.EXPECT().Tx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context, domain.UnitOfWork) error) error {
		return fn(ctx, uow)
	}).Times(1)

	sut := DataService{unitOfWork: uow}

	newData, err := sut.Upload(ctx, data)

	assert.NoError(t, err)
	assert.NotNil(t, newData)
}

// TestDataService_UploadShouldUpdateDataSuccessfully просто изменяем данные(если данные изменились версия должна быть больше той, чем имеем)
func TestDataService_UploadShouldUpdateDataSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	id, data, local := generateData("some pass", domain.LoginPassType)
	uow := mocks.NewMockUnitOfWork(ctrl)
	mockRep := mocks.NewMockDataRepository(ctrl)
	mockRep.EXPECT().Get(ctx, id).Return(local, nil).Times(2)
	mockRep.EXPECT().Update(ctx, local).Return(nil)
	uow.EXPECT().DataRepository().Return(mockRep).Times(2)
	uow.EXPECT().Tx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context, domain.UnitOfWork) error) error {
		return fn(ctx, uow)
	}).Times(1)

	sut := DataService{unitOfWork: uow}

	newData, err := sut.Upload(ctx, data)

	assert.NoError(t, err)
	assert.NotNil(t, newData)
	assert.Equal(t, int32(2), newData.Version)
}

// TestDataService_UploadShouldBeWithConflict проверяем что произойдет конфликт если версия будет равна или меньше текущей
func TestDataService_UploadShouldBeWithConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	id, data, local := generateData("some pass", domain.LoginPassType)
	local.Version = 2
	uow := mocks.NewMockUnitOfWork(ctrl)
	mockRep := mocks.NewMockDataRepository(ctrl)
	mockRep.EXPECT().Get(ctx, id).Return(local, nil).Times(1)
	mockRep.EXPECT().Update(ctx, local).Return(nil).Times(0)
	uow.EXPECT().DataRepository().Return(mockRep).Times(1)
	uow.EXPECT().Tx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context, domain.UnitOfWork) error) error {
		return fn(ctx, uow)
	}).Times(1)

	sut := DataService{unitOfWork: uow}

	newData, err := sut.Upload(ctx, data)

	assert.ErrorIs(t, err, domain.ErrDataConflict)
	assert.Nil(t, newData)
}

func TestDataService_BatchUploadUpdateDataSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	index := 5
	ids := make([]guid.Guid, index)
	data := make([]*Data, index)
	localData := make([]*domain.Data, index)
	for i := 0; i < index; i++ {
		id, dt, local := generateData(fmt.Sprintf("some pass %d", i+1), domain.LoginPassType)
		ids[i] = id
		data[i] = dt
		localData[i] = local
	}

	uow := mocks.NewMockUnitOfWork(ctrl)
	mockRep := mocks.NewMockDataRepository(ctrl)
	uow.EXPECT().DataRepository().Return(mockRep).Times(2)
	for i := 0; i < index; i++ {
		mockRep.EXPECT().Get(ctx, ids[i]).Return(localData[i], nil).Times(2)
		mockRep.EXPECT().Update(ctx, localData[i]).Return(nil)
	}

	uow.EXPECT().Tx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context, domain.UnitOfWork) error) error {
		return fn(ctx, uow)
	}).Times(1)

	sut := DataService{unitOfWork: uow}

	newData, err := sut.BatchUpload(ctx, data)

	assert.NoError(t, err)
	assert.NotNil(t, newData)
	for _, i := range newData {
		assert.Equal(t, int32(2), i.Version)
	}
}

func generateData(name string, tt domain.DataType) (id guid.Guid, data *Data, localData *domain.Data) {
	id = *guid.New()
	dek, dekNonce := make([]byte, 32), make([]byte, 16)
	_, _ = rand.Read(dek)
	_, _ = rand.Read(dekNonce)
	dt, dtNonce := make([]byte, 4026), make([]byte, 16)
	_, _ = rand.Read(dt)
	_, _ = rand.Read(dtNonce)
	data = &Data{
		ID:           id,
		Name:         name,
		Dek:          dek,
		Large:        false,
		Type:         tt,
		DekNonce:     dekNonce,
		Payload:      dt,
		PayloadNonce: dtNonce,
		Version:      1,
	}
	localData = &domain.Data{
		ID:           id,
		CreatedAt:    time.Now(),
		Name:         name,
		Dek:          dek,
		Large:        false,
		Type:         tt,
		DekNonce:     dekNonce,
		Payload:      dt,
		PayloadNonce: dtNonce,
		Version:      1,
	}
	return
}
