package usecase

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/infrastructure/persistence"
	"github.com/DimKa163/keeper/internal/server/shared/auth"
	"github.com/beevik/guid"
)

var syncTypeName = reflect.TypeOf(domain.Secret{}).Name()
var (
	ErrVersionConflict = errors.New("conflict")
)

type OperationType int

const (
	DefaultOperation OperationType = iota
	BeginOperation
	ChunkOperation
	EndOperation
)

type (
	Secret struct {
		ID         guid.Guid
		ModifiedAt time.Time
		Type       domain.SecretType
		Dek        []byte
		Data       []byte
		Version    int32
		Deleted    bool
	}
	Push struct {
		Secret *Secret
		Type   OperationType
		Buffer []byte
	}
)

type SyncService struct {
	uow domain.UnitOfWork
	fp  domain.Filer
}

func NewSyncService(uow domain.UnitOfWork, fp domain.Filer) *SyncService {
	return &SyncService{uow: uow, fp: fp}
}

func (ss *SyncService) ValidateVersion(ctx context.Context, version int32) error {
	userID, err := auth.User(ctx)
	if err != nil {
		return err
	}
	syncStateRepository := ss.uow.SyncStateRepository()
	syncState, err := syncStateRepository.Get(ctx, syncTypeName, userID)
	if err != nil {
		return err
	}
	if syncState.Value > version {
		return ErrVersionConflict
	}
	return nil
}

func (ss *SyncService) File(id guid.Guid, version int32) (io.ReadCloser, error) {
	return ss.fp.OpenRead(id.String(), version)
}

func (ss *SyncService) Push(ctx context.Context, fn func(ctx context.Context) (*Push, error)) error {
	if err := ss.uow.Tx(ctx, func(ctx context.Context, work domain.UnitOfWork) error {
		userID, err := auth.User(ctx)
		if err != nil {
			return err
		}
		syncStateRepository := work.SyncStateRepository()
		syncState, err := syncStateRepository.Get(ctx, syncTypeName, userID)
		syncState.Value += 1
		if err != nil {
			return err
		}
		for {
			req, err := fn(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			switch req.Type {
			case DefaultOperation:
				if err = ss.push(ctx, work, syncState, req); err != nil {
					return err
				}
			case BeginOperation:
				if err = ss.startUploadFile(ctx, work, syncState, req); err != nil {
					return err
				}
			case ChunkOperation:
				if err = ss.writeChunk(ctx, work, syncState, req); err != nil {
					return err
				}
			case EndOperation:
				if err = ss.endFile(ctx, work, syncState, req); err != nil {
					return err
				}
			}
		}
		return syncStateRepository.Update(ctx, syncState)
	}); err != nil {
		return err
	}
	return nil
}
func (ss *SyncService) Poll(ctx context.Context, since int32) ([]*domain.Secret, int32, error) {
	user, err := auth.User(ctx)
	if err != nil {
		return nil, -1, err
	}
	rep := ss.uow.SecretRepository()
	data, err := rep.GetAll(ctx, user, since)
	if err != nil {
		return nil, -1, err
	}
	userID, err := auth.User(ctx)
	if err != nil {
		return nil, -1, err
	}
	syncRepository := ss.uow.SyncStateRepository()
	state, err := syncRepository.Get(ctx, syncTypeName, userID)
	if err != nil {
		if !errors.Is(err, persistence.ErrResourceNotFound) {
			return nil, -1, err
		}
		state = &domain.SyncState{
			ID:    syncTypeName,
			Value: 0,
		}
	}
	return data, state.Value, nil
}

func (ss *SyncService) push(ctx context.Context, uow domain.UnitOfWork, state *domain.SyncState, p *Push) error {
	secret := p.Secret
	dataRepository := uow.SecretRepository()
	data, err := dataRepository.Get(ctx, secret.ID)
	if err != nil && !errors.Is(err, persistence.ErrResourceNotFound) {
		return err
	}
	if data != nil {
		data.ModifiedAt = secret.ModifiedAt
		data.Dek = secret.Dek
		data.Payload = secret.Data
		data.Version = state.Value
		data.Deleted = secret.Deleted
		return dataRepository.Update(ctx, data)
	}
	userID, err := auth.User(ctx)
	if err != nil {
		return err
	}
	data = &domain.Secret{
		ID:         secret.ID,
		ModifiedAt: secret.ModifiedAt,
		UserID:     userID,
		Type:       secret.Type,
		BigData:    false,
		Dek:        secret.Dek,
		Payload:    secret.Data,
		Deleted:    secret.Deleted,
		Version:    state.Value,
	}
	return dataRepository.Insert(ctx, data)
}

func (ss *SyncService) startUploadFile(ctx context.Context, uow domain.UnitOfWork, state *domain.SyncState, p *Push) error {
	secret := p.Secret
	secretRep := uow.SecretRepository()
	data, err := secretRep.Get(ctx, secret.ID)
	if err != nil && !errors.Is(err, persistence.ErrResourceNotFound) {
		return err
	}
	if data != nil {
		data.Deleted = secret.Deleted
		if data.Deleted {
			if err = ss.fp.Remove(data.ID.String(), data.Version); err != nil {
				return err
			}
			data.ModifiedAt = secret.ModifiedAt
			data.Version = state.Value
		}
		return secretRep.Update(ctx, data)
	}
	userID, err := auth.User(ctx)
	if err != nil {
		return err
	}
	data = &domain.Secret{
		ID:         secret.ID,
		ModifiedAt: secret.ModifiedAt,
		UserID:     userID,
		Type:       secret.Type,
		BigData:    true,
	}
	return secretRep.Insert(ctx, data)
}

func (ss *SyncService) writeChunk(ctx context.Context, uow domain.UnitOfWork, state *domain.SyncState, p *Push) error {
	secret := p.Secret
	dataRepository := uow.SecretRepository()
	data, err := dataRepository.Get(ctx, secret.ID)
	if err != nil {
		return err
	}
	f, err := ss.fp.OpenWrite(data.ID.String(), state.Value)
	if err != nil {
		return err
	}
	defer func(f io.WriteCloser) {
		_ = f.Close()
	}(f)
	_, err = f.Write(p.Buffer)
	if err != nil {
		return err
	}
	return nil
}

func (ss *SyncService) endFile(ctx context.Context, uow domain.UnitOfWork, state *domain.SyncState, p *Push) error {
	secret := p.Secret
	dataRepository := uow.SecretRepository()
	data, err := dataRepository.Get(ctx, secret.ID)
	if err != nil {
		return err
	}
	oldVersion := data.Version
	data.Dek = secret.Dek
	data.Payload = secret.Data
	data.Version = state.Value
	data.ModifiedAt = secret.ModifiedAt
	if err = dataRepository.Update(ctx, data); err != nil {
		return err
	}
	// удаляем старую версию
	if err = ss.fp.Remove(data.ID.String(), oldVersion); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
