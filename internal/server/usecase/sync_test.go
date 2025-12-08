package usecase

import (
	"context"
	"crypto/rand"
	"github.com/DimKa163/keeper/internal/mocks"
	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/infrastructure/persistence"
	"github.com/DimKa163/keeper/internal/server/shared/auth"
	"github.com/DimKa163/keeper/internal/shared"
	"github.com/beevik/guid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"io/fs"
	"testing"
	"time"
)

func TestSyncService_Push_DefaultShouldBeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	userId := *guid.New()
	ctx = auth.SetUser(ctx, userId)
	dek := make([]byte, 32)
	_, _ = rand.Read(dek)
	data := make([]byte, shared.KB)
	_, _ = rand.Read(data)
	message := &Push{
		Type: DefaultOperation,
		Secret: &Secret{
			ID:         *guid.New(),
			ModifiedAt: time.Now(),
			Type:       domain.LoginPassType,
			Deleted:    false,
			Data:       data,
			Dek:        dek,
			Version:    1,
		},
	}
	state := &domain.SyncState{
		ID:    syncTypeName,
		Value: 0,
	}
	txUow := mocks.NewMockUnitOfWork(ctrl)
	uow := newMockUow(txUow)
	syncRepository := mocks.NewMockSyncStateRepository(ctrl)
	syncRepository.EXPECT().Get(ctx, syncTypeName, userId).Return(state, nil)
	syncRepository.EXPECT().Update(ctx, &domain.SyncState{
		ID:    syncTypeName,
		Value: 1,
	}).Return(nil)
	secretRepository := mocks.NewMockSecretRepository(ctrl)
	secretRepository.EXPECT().Get(ctx, message.Secret.ID).Return(nil, persistence.ErrResourceNotFound)
	secretRepository.EXPECT().Insert(ctx, gomock.Any()).Return(nil)
	txUow.EXPECT().SyncStateRepository().Return(syncRepository)
	txUow.EXPECT().SecretRepository().Return(secretRepository)
	mockFiler := mocks.NewMockFiler(ctrl)

	syncService := NewSyncService(uow, mockFiler)
	arr := make([]*Push, 1)
	arr[0] = message
	str := newMockStream(arr)
	err := syncService.Push(ctx, str.Next)

	assert.NoError(t, err)
	assert.Equal(t, int32(1), state.Value)
}

func TestSyncService_Push_FileShouldBeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	userId := *guid.New()
	ctx = auth.SetUser(ctx, userId)
	dek := make([]byte, 32)
	_, _ = rand.Read(dek)
	data := make([]byte, shared.KB)
	_, _ = rand.Read(data)

	msgs := make([]*Push, 3)
	id := *guid.New()
	msgs[0] = &Push{
		Type: BeginOperation,
		Secret: &Secret{
			ID:         id,
			ModifiedAt: time.Now(),
			Version:    1,
		},
	}
	buffer := make([]byte, shared.MB)
	_, _ = rand.Read(buffer)
	msgs[1] = &Push{
		Type: ChunkOperation,
		Secret: &Secret{
			ID: id,
		},
		Buffer: buffer,
	}
	msgs[2] = &Push{
		Type: EndOperation,
		Secret: &Secret{
			ID:         id,
			ModifiedAt: time.Now(),
			Version:    1,
			Dek:        dek,
			Data:       data,
			Deleted:    false,
			Type:       domain.OtherType,
		},
	}
	state := &domain.SyncState{
		ID:    syncTypeName,
		Value: 0,
	}
	txUow := mocks.NewMockUnitOfWork(ctrl)
	uow := newMockUow(txUow)
	syncRepository := mocks.NewMockSyncStateRepository(ctrl)
	syncRepository.EXPECT().Get(ctx, syncTypeName, userId).Return(state, nil)
	syncRepository.EXPECT().Update(ctx, &domain.SyncState{
		ID:    syncTypeName,
		Value: 1,
	}).Return(nil)
	secretRepository := mocks.NewMockSecretRepository(ctrl)
	secretRepository.EXPECT().Get(ctx, id).Return(nil, persistence.ErrResourceNotFound).Times(1)
	secretRepository.EXPECT().Get(ctx, id).Return(&domain.Secret{
		ID:         id,
		CreatedAt:  time.Now(),
		ModifiedAt: msgs[0].Secret.ModifiedAt,
	}, nil).Times(len(msgs) - 1)
	secretRepository.EXPECT().Insert(ctx, gomock.Any()).Return(nil)
	secretRepository.EXPECT().Update(ctx, gomock.Any()).Return(nil)
	txUow.EXPECT().SyncStateRepository().Return(syncRepository)
	txUow.EXPECT().SecretRepository().Return(secretRepository).Times(len(msgs))
	mockFiler := mocks.NewMockFiler(ctrl)
	mockFiler.EXPECT().OpenWrite(id.String(), state.Value+1).Return(&mockWriterCloser{}, nil).Times(len(msgs) - 2)
	mockFiler.EXPECT().Remove(id.String(), int32(0)).Return(fs.ErrNotExist)
	syncService := NewSyncService(uow, mockFiler)

	str := newMockStream(msgs)

	err := syncService.Push(ctx, str.Next)
	assert.NoError(t, err)
}

type mockUow struct {
	tx domain.UnitOfWork
}

func (m *mockUow) UserRepository() domain.UserRepository {
	return m.tx.UserRepository()
}

func (m *mockUow) SecretRepository() domain.SecretRepository {
	return m.tx.SecretRepository()
}

func (m *mockUow) SyncStateRepository() domain.SyncStateRepository {
	return m.tx.SyncStateRepository()
}

func (m *mockUow) Tx(ctx context.Context, fn func(ctx context.Context, work domain.UnitOfWork) error) error {
	return fn(ctx, m.tx)
}

func newMockUow(tx domain.UnitOfWork) domain.UnitOfWork {
	return &mockUow{tx: tx}
}

type mockStream struct {
	items []*Push
	i     int
}

func newMockStream(items []*Push) *mockStream {
	return &mockStream{items: items, i: 0}
}

func (m *mockStream) Next(_ context.Context) (*Push, error) {
	defer func() { m.i++ }()
	if m.i >= len(m.items) {
		return nil, io.EOF
	}
	return m.items[m.i], nil
}

type mockWriterCloser struct {
}

func (m *mockWriterCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}
func (m *mockWriterCloser) Close() error {
	return nil
}
