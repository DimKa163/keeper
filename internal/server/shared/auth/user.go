// Package auth
package auth

import (
	"context"
	"errors"

	"github.com/beevik/guid"
)

var ErrUserNotFound = errors.New("user not found in context")

type UserID string

const (
	user UserID = "userID"
)

// User get user id from context
func User(ctx context.Context) (guid.Guid, error) {
	userID, ok := ctx.Value(user).(guid.Guid)
	if !ok {
		return guid.Guid{}, ErrUserNotFound
	}
	return userID, nil
}

// SetUser set user id to context
func SetUser(ctx context.Context, userID guid.Guid) context.Context {
	ctx = context.WithValue(ctx, user, userID)
	return ctx
}
