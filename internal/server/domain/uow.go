package domain

import "context"

// UnitOfWork provider request to storage
type UnitOfWork interface {
	UserRepository() UserRepository

	SecretRepository() SecretRepository

	SyncStateRepository() SyncStateRepository

	Tx(ctx context.Context, fn func(ctx context.Context, work UnitOfWork) error) error
}
