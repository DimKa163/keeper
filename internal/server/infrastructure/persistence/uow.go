package persistence

import (
	"context"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/DimKa163/keeper/internal/server/shared/db"
)

type UnitOfWork struct {
	db db.QueryExecutor
}

func NewUnitOfWork(db db.QueryExecutor) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) UserRepository() domain.UserRepository {
	return NewUserRepository(u.db)
}

func (u *UnitOfWork) SecretRepository() domain.SecretRepository {
	return NewSecretRepository(u.db)
}

func (u *UnitOfWork) SyncStateRepository() domain.SyncStateRepository {
	return NewSyncStateRepository(u.db)
}

func (u *UnitOfWork) Tx(ctx context.Context, fn func(ctx context.Context, work domain.UnitOfWork) error) error {
	var err error
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	if err = fn(ctx, NewUnitOfWork(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
