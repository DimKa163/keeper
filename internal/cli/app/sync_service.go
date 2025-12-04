package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/DimKa163/keeper/internal/pb"
	"github.com/DimKa163/keeper/internal/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"reflect"
)

var ErrCorruptedData = errors.New("corrupted data")

var ErrConflictData = errors.New("conflict detected! pull first")
var syncTypeName = reflect.TypeOf(core.Record{}).Name()

type (
	SyncOption struct {
		PushOnly bool
		PullOnly bool
		Force    bool
	}
)

type Syncer interface {
	Sync(ctx context.Context, option *SyncOption) error
}

type SyncService struct {
	client       *RemoteClient
	db           *sql.DB
	fileProvider *shared.FileProvider
}

func NewSyncService(
	client *RemoteClient,
	db *sql.DB,
	fileProvider *shared.FileProvider,
) *SyncService {
	return &SyncService{
		client:       client,
		db:           db,
		fileProvider: fileProvider,
	}
}

// Sync синхронизация данных между клиентом и сервером
func (ss *SyncService) Sync(ctx context.Context, option *SyncOption) error {
	var err error
	tx, err := ss.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()
	fmt.Println("starting sync process")
	syncState, err := getState(ctx, tx)
	if err != nil {
		return err
	}
	if !option.PullOnly && !option.PushOnly {
		if err = ss.push(ctx, tx, syncState, option.Force); err != nil {
			if errors.Is(err, ErrConflictData) {
				fmt.Println("pull first, conflict detected")
			}
			return err
		}
		if err = ss.pull(ctx, tx, syncState, option.Force); err != nil {
			return err
		}

	} else if option.PushOnly {
		if err = ss.push(ctx, tx, syncState, option.Force); err != nil {
			if errors.Is(err, ErrConflictData) {
				fmt.Println("pull first, conflict detected")
			}
			return err
		}
	} else {
		if err = ss.pull(ctx, tx, syncState, option.Force); err != nil {
			return err
		}
	}
	return nil
}

func (ss *SyncService) push(ctx context.Context, tx *sql.Tx, syncState *core.SyncState, force bool) error {
	var err error
	fmt.Printf("current version: %d\n", syncState.Value)
	records, err := persistence.TxGetAllRecordGreater(ctx, tx, syncState.Value)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}
	secrets := make([]*pb.Secret, len(records))

	var request pb.PushRequest
	request.SetSecrets(secrets)
	request.SetForce(force)
	for i, record := range records {
		secrets[i] = toSecret(record)
	}
	_, err = ss.client.Push(ctx, &request)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.FailedPrecondition:
				return ErrConflictData
			}
		}
		return err
	}
	for _, record := range records {
		if !record.BigData {
			continue
		}
		if err = ss.sendFile(ctx, record); err != nil {
		}
	}
	return nil
}

func (ss *SyncService) sendFile(ctx context.Context, record *core.Record) error {
	stream, err := ss.client.PushStream(ctx)
	if err != nil {
		return err
	}
	file, err := ss.fileProvider.OpenRead(record.ID, record.Version)
	if err != nil {
		return err
	}
	defer func(file io.ReadCloser) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed to close file: %s\n", err)
		}
	}(file)
	buffer := make([]byte, shared.MB)
	for {
		var chunk pb.Chunk
		n, err := file.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		chunk.SetType(pb.ChunkType_FilePart)
		chunk.SetId(record.ID)
		chunk.SetBuffer(buffer[:n])
		if err = stream.Send(&chunk); err != nil {
			return err
		}
	}
	var chunk pb.Chunk
	chunk.SetId(record.ID)
	chunk.SetSecret(toSecret(record))
	chunk.SetType(pb.ChunkType_EndData)
	if err = stream.Send(&chunk); err != nil {
		return err
	}
	if _, err = stream.CloseAndRecv(); err != nil {
		return err
	}
	return nil
}

func (ss *SyncService) pull(ctx context.Context, tx *sql.Tx, syncState *core.SyncState, force bool) error {
	fmt.Println("starting receiving secrets from server")
	var err error
	var request pb.PullRequest
	request.SetSince(syncState.Value)
	resp, err := ss.client.Pull(ctx, &request)
	if err != nil {
		return err
	}
	var hasConflict bool
	for _, item := range resp.GetSecrets() {
		record, err := persistence.TxGetRecordByID(ctx, tx, item.GetId())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if record != nil {
			if !record.ModifiedAt.Equal(item.GetModifiedAt().AsTime()) && record.Version <= item.GetVersion() && !force {
				hasConflict = true
				fmt.Printf("detected conflict for secret %s\n", item.GetId())
				if err = ss.createConflict(ctx, tx, record, item); err != nil {
					return err
				}
				continue
			}
			if err = ss.update(ctx, tx, record, item); err != nil {
				return err
			}
		} else {
			if err = ss.create(ctx, tx, item); err != nil {
				return err
			}
		}
	}
	if hasConflict {
		count, err := persistence.TxGetConflictCount(ctx, tx)
		if err != nil {
			return err
		}
		fmt.Printf("❗😠detected conflict. count %d\n", count)
		return nil
	}
	syncState.Value = resp.GetVersion()
	if err = persistence.SaveState(ctx, tx, syncState); err != nil {
		return err
	}
	fmt.Println("✅ secrets received successfully")
	return nil
}

func (ss *SyncService) createConflict(
	ctx context.Context,
	tx *sql.Tx,
	target *core.Record,
	secret *pb.Secret,
) error {
	if secret.GetIsBig() {
		var pollStream pb.PullStreamRequest
		pollStream.SetId(secret.GetId())
		pollStream.SetVersion(secret.GetVersion())
		stream, err := ss.client.PullStream(ctx, &pollStream)
		if err != nil {
			return err
		}
		file, err := ss.fileProvider.OpenWrite(secret.GetId(), secret.GetVersion(), "remote")
		if err != nil {
			return err
		}
		defer func(file io.WriteCloser) {
			err := file.Close()
			if err != nil {
				fmt.Printf("failed to close file: %s\n", err)
			}
		}(file)
		for {
			chunk, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
			}
			switch chunk.GetType() {
			case pb.ChunkType_FilePart:
				buffer := chunk.GetBuffer()
				written := 0
				for written < len(buffer) {
					n, err := file.Write(buffer[written:])
					if err != nil {
						return err
					}
					written += n
				}
			case pb.ChunkType_ErrData:
				continue
			}
		}
	}
	conflict := &core.Conflict{
		RecordID: target.ID,
		Local: &core.ConflictItem{
			Record:  target,
			Deleted: target.Deleted,
		},
		Remote: &core.ConflictItem{
			Record:  toRecord(secret),
			Deleted: secret.GetDeleted(),
		},
	}
	return persistence.TxInsertConflict(ctx, tx, conflict)
}

func (ss *SyncService) create(ctx context.Context, tx *sql.Tx, secret *pb.Secret) error {
	record := toRecord(secret)
	return persistence.TxInsertRecord(ctx, tx, record)
}

func (ss *SyncService) createFile(ctx context.Context, tx *sql.Tx, secret *pb.Secret) error {
	var pollStream pb.PullStreamRequest
	pollStream.SetId(secret.GetId())
	pollStream.SetVersion(secret.GetVersion())
	stream, err := ss.client.PullStream(ctx, &pollStream)
	if err != nil {
		return err
	}
	file, err := ss.fileProvider.OpenWrite(secret.GetId(), secret.GetVersion())
	if err != nil {
		return err
	}
	defer func(file io.WriteCloser) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed to close file: %s\n", err)
		}
	}(file)
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		switch chunk.GetType() {
		case pb.ChunkType_FilePart:
			buffer := chunk.GetBuffer()
			written := 0
			for written < len(buffer) {
				n, err := file.Write(buffer[written:])
				if err != nil {
					return err
				}
				written += n
			}
		case pb.ChunkType_ErrData:
			continue
		}
	}
	record := toRecord(secret)
	return persistence.TxInsertRecord(ctx, tx, record)
}
func (ss *SyncService) update(ctx context.Context, tx *sql.Tx, target *core.Record, secret *pb.Secret) error {
	if secret.GetIsBig() {
		if err := ss.updateFile(ctx, tx, target, secret); err != nil {
			return err
		}
	}
	target.Deleted = secret.GetDeleted()
	target.ModifiedAt = secret.GetModifiedAt().AsTime()
	target.Dek = secret.GetDek()
	target.Data = secret.GetData()
	target.Version = secret.GetVersion()
	return persistence.TxUpdateRecord(ctx, tx, target)
}

func (ss *SyncService) updateFile(ctx context.Context, tx *sql.Tx, target *core.Record, secret *pb.Secret) error {
	if target.Version < secret.GetVersion() {
		var pollStream pb.PullStreamRequest
		pollStream.SetId(target.ID)
		pollStream.SetVersion(secret.GetVersion())
		stream, err := ss.client.PullStream(ctx, &pollStream)
		if err != nil {
			return err
		}
		file, err := ss.fileProvider.OpenWrite(target.ID, secret.GetVersion())
		if err != nil {
			return err
		}
		defer func(file io.WriteCloser) {
			err := file.Close()
			if err != nil {
				fmt.Printf("failed to close file: %s\n", err)
			}
		}(file)
		for {
			chunk, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			switch chunk.GetType() {
			case pb.ChunkType_FilePart:
				buffer := chunk.GetBuffer()
				written := 0
				for written < len(buffer) {
					n, err := file.Write(buffer[written:])
					if err != nil {
						return err
					}
					written += n
				}
			case pb.ChunkType_ErrData:
				return ss.handleCorruptData(ctx, tx, target.ID)
			}
		}
		if err = ss.fileProvider.Remove(target.ID, target.Version); err != nil {
			return err
		}
	}
	return nil
	//
	//target.DekNonce = secret.GetDekNonce()
	//target.Dek = secret.GetDek()
	//target.DataNonce = secret.GetDataNonce()
	//target.Data = secret.GetData()
	//target.FileNonce = secret.GetFileDataNonce()
	//target.Version = secret.GetVersion()
	//return persistence.TxUpdateRecord(ctx, tx, target)
}

func (ss *SyncService) handleCorruptData(ctx context.Context, tx *sql.Tx, id string) error {
	return persistence.TxCorruptRecord(ctx, tx, true, id)
}

func getState(ctx context.Context, tx *sql.Tx) (*core.SyncState, error) {
	syncState, err := persistence.TxGetState(ctx, tx, syncTypeName)
	if err != nil {
		if !errors.Is(sql.ErrNoRows, err) {
			return nil, err
		}
		syncState = &core.SyncState{
			ID:    syncTypeName,
			Value: 0,
		}
	}
	return syncState, nil
}
func toSecret(record *core.Record) *pb.Secret {
	var secret pb.Secret
	secret.SetId(record.ID)
	secret.SetModifiedAt(timestamppb.New(record.ModifiedAt))
	secret.SetDek(record.Dek)
	secret.SetData(record.Data)
	secret.SetIsBig(record.BigData)
	secret.SetVersion(record.Version)
	secret.SetDeleted(record.Deleted)
	switch record.Type {
	case core.LoginPassType:
		secret.SetType(pb.Secret_LoginPass)
	case core.TextType:
		secret.SetType(pb.Secret_Text)
	case core.BankCardType:
		secret.SetType(pb.Secret_BankCard)
	case core.OtherType:
		secret.SetType(pb.Secret_Other)
	}
	return &secret
}

func toRecord(secret *pb.Secret) *core.Record {
	var record core.Record
	record.ID = secret.GetId()
	record.BigData = secret.GetIsBig()
	record.Dek = secret.GetDek()
	record.Data = secret.GetData()
	record.Version = secret.GetVersion()
	switch secret.GetType() {
	case pb.Secret_LoginPass:
		record.Type = core.LoginPassType
	case pb.Secret_Text:
		record.Type = core.TextType
	case pb.Secret_BankCard:
		record.Type = core.BankCardType
	case pb.Secret_Other:
		record.Type = core.OtherType
	}
	return &record
}
