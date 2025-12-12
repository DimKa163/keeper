// Package common for context.Context
package common

import (
	"context"
	"errors"
)

const stateName = "recordState"

const (
	key        MasterKey = "github.com/DimKa163/keeper_MasterKey"
	hostName   HostName  = "github.com/DimKa163/keeper_Hostname"
	versionKey Version   = "github.com/DimKa163/keeper_version"
)

type MasterKey string

type HostName string

type Version string

var (
	ErrMasterKeyNotRegistered = errors.New("keeper: master key not registered")
)

// SetHostName установить название машины
func SetHostName(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, hostName, value)
}

// SetMasterKey установить значение мастер ключа
func SetMasterKey(ctx context.Context, masterKey []byte) context.Context {
	return context.WithValue(ctx, key, masterKey)
}

// GetMasterKey получить мастер ключ
func GetMasterKey(ctx context.Context) ([]byte, error) {
	if v, ok := ctx.Value(key).([]byte); ok {
		return v, nil
	}
	return nil, ErrMasterKeyNotRegistered
}

// SetVersion установить версию
func SetVersion(ctx context.Context, version int32) context.Context {
	return context.WithValue(ctx, versionKey, version)
}

// GetVersion получить версию
func GetVersion(ctx context.Context) int32 {
	if v, ok := ctx.Value(versionKey).(int32); ok {
		return v
	}
	return -1
}
