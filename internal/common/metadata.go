package common

import (
	"context"
	"errors"
	"strconv"

	"google.golang.org/grpc/metadata"
)

var (
	ErrMetadataNotFound     = errors.New("metadata not found")
	ErrClientVersionMissing = errors.New("client version missing")
)

const (
	ClientVERSION = `client_version`
	FORCE         = `force_update`
)

func WriteClientVersion(ctx context.Context, version int32) context.Context {
	return metadata.AppendToOutgoingContext(ctx, ClientVERSION, strconv.Itoa(int(version)))
}

func WriteForce(ctx context.Context, force bool) context.Context {
	return metadata.AppendToOutgoingContext(ctx, FORCE, strconv.FormatBool(force))
}

func ReadVersionFromHeader(ctx context.Context) (int32, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return -1, ErrMetadataNotFound
	}
	if clientVersion, ok := md[ClientVERSION]; ok {
		v, err := strconv.Atoi(clientVersion[0])
		if err != nil {
			return -1, err
		}
		return int32(v), nil
	}
	return -1, ErrClientVersionMissing
}

func ReadForceFromHeader(ctx context.Context) (bool, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false, ErrMetadataNotFound
	}
	if forceValue, ok := md[FORCE]; ok {
		v, err := strconv.ParseBool(forceValue[0])
		if err != nil {
			return false, err
		}
		return v, nil
	}
	return false, ErrClientVersionMissing
}
