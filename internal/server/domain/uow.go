package domain

import "context"

// UnitOfWork provider request to storage
type UnitOfWork interface {
	UserRepository() UserRepository

	Tx(ctx context.Context, fn func(ctx context.Context, work UnitOfWork) error) error
}
