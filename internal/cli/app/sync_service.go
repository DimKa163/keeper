package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/DimKa163/keeper/internal/common"
	"github.com/DimKa163/keeper/internal/datatool"
	"github.com/DimKa163/keeper/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

type PushSecretStream grpc.ClientStreamingClient[pb.PushOperation, pb.PushResponse]

type SyncService struct {
	client       *RemoteClient
	db           *sql.DB
	fileProvider *datatool.FileProvider
}

func NewSyncService(
	client *RemoteClient,
	db *sql.DB,
	fileProvider *datatool.FileProvider,
) *SyncService {
	return &SyncService{
		client:       client,
		db:           db,
		fileProvider: fileProvider,
	}
}

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
	ctx = common.WriteClientVersion(ctx, syncState.Value)
	ctx = common.WriteForce(ctx, force)
	stream, err := ss.client.SyncClient.PushStream(ctx)
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.BigData {
			if err = ss.pushFile(stream, record); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
			}
			continue
		}
		op := toDefault(record)
		if err = stream.Send(op); err != nil {
			return err
		}
	}
	if _, err = stream.CloseAndRecv(); err != nil {
		return err
	}
	return nil
}

func (ss *SyncService) pushFile(stream PushSecretStream, record *core.Record) error {
	var n int
	begin := toBegin(record)
	if err := stream.Send(begin); err != nil {
		return err
	}
	reader, err := ss.fileProvider.OpenRead(record.ID, record.Version)
	if err != nil {
		return err
	}
	defer func(reader io.ReadCloser) {
		err = reader.Close()
		if err != nil {
			fmt.Printf("failed to close file: %s\n", err)
		}
	}(reader)
	buffer := make([]byte, datatool.MB)
	for {
		n, err = reader.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		msg := toChunk(record, buffer, n)
		if err = stream.Send(msg); err != nil {
			return err
		}
	}
	end := toEndFile(record)
	if err = stream.Send(end); err != nil {
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
		var record *core.Record
		record, err = persistence.TxGetRecordByID(ctx, tx, item.GetId())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if record != nil {
			if isConflictDetected(record, item, syncState) && !force {
				hasConflict = true
				fmt.Printf("detected conflict for secret %s\n", item.GetId())
				if err = ss.createConflict(ctx, tx, record, item); err != nil {
					return err
				}
				continue
			}
			if item.GetDeleted() {
				if err = ss.delete(ctx, tx, record); err != nil {
					return err
				}
				continue
			}
			if err = ss.update(ctx, tx, record, item); err != nil {
				return err
			}
		} else {
			if item.GetDeleted() {
				continue
			}
			if err = ss.create(ctx, tx, item); err != nil {
				return err
			}
		}
	}
	if hasConflict {
		var count int64
		count, err = persistence.TxGetConflictCount(ctx, tx)
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
	if secret.GetIsBig() && !secret.GetDeleted() {
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
func (ss *SyncService) update(ctx context.Context, tx *sql.Tx, target *core.Record, secret *pb.Secret) error {
	if secret.GetIsBig() {
		if err := ss.updateFile(ctx, target, secret); err != nil {
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

func (ss *SyncService) updateFile(ctx context.Context, target *core.Record, secret *pb.Secret) error {
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
			err = file.Close()
			if err != nil {
				fmt.Printf("failed to close file: %s\n", err)
			}
		}(file)
		for {
			var chunk *pb.Chunk
			chunk, err = stream.Recv()
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
					var n int
					n, err = file.Write(buffer[written:])
					if err != nil {
						return err
					}
					written += n
				}
			case pb.ChunkType_ErrData:
				return nil
			}
		}
		if target.BigData {
			if err = ss.fileProvider.Remove(target.ID, target.Version); err != nil {
				return err
			}
		}
	}
	return nil
}
func (ss *SyncService) create(ctx context.Context, tx *sql.Tx, secret *pb.Secret) error {
	record := toRecord(secret)
	if record.BigData {
		if err := ss.createFile(ctx, secret); err != nil {
			return err
		}
	}
	return persistence.TxInsertRecord(ctx, tx, record)
}

func (ss *SyncService) createFile(ctx context.Context, secret *pb.Secret) error {
	var pollStream pb.PullStreamRequest
	pollStream.SetId(secret.GetId())
	pollStream.SetVersion(secret.GetVersion())
	stream, err := ss.client.PullStream(ctx, &pollStream)
	if err != nil {
		return err
	}
	if err = ss.fileProvider.Remove(secret.GetId(), secret.GetVersion()); err != nil && !os.IsNotExist(err) {
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
	return nil
}

func (ss *SyncService) delete(ctx context.Context, tx *sql.Tx, target *core.Record) error {
	if target.BigData {
		if err := ss.fileProvider.Remove(target.ID, target.Version); err != nil {
			return err
		}
	}
	return persistence.TxDeleteRecord(ctx, tx, target.ID)
}

func isConflictDetected(record *core.Record, secret *pb.Secret, syncState *core.SyncState) bool {
	return record.IsChanged(syncState) && !record.ModifiedAt.Equal(secret.GetModifiedAt().AsTime())
}

func toDefault(record *core.Record) *pb.PushOperation {
	var op pb.PushOperation
	op.SetType(pb.OperationType_Default)
	op.SetSecret(toSecret(record))
	return &op
}
func toBegin(record *core.Record) *pb.PushOperation {
	var op pb.PushOperation
	op.SetType(pb.OperationType_Begin)
	var secret pb.Secret
	secret.SetId(record.ID)
	secret.SetModifiedAt(timestamppb.New(record.ModifiedAt))
	secret.SetVersion(record.Version)
	op.SetSecret(&secret)
	return &op
}

func toChunk(record *core.Record, buffer []byte, n int) *pb.PushOperation {
	var op pb.PushOperation
	op.SetType(pb.OperationType_BinaryPart)
	var secret pb.Secret
	secret.SetId(record.ID)
	op.SetSecret(&secret)
	op.SetBuffer(buffer[:n])
	return &op
}

func toEndFile(record *core.Record) *pb.PushOperation {
	var op pb.PushOperation
	op.SetType(pb.OperationType_End)
	var secret pb.Secret
	secret.SetId(record.ID)
	secret.SetModifiedAt(timestamppb.New(record.ModifiedAt))
	secret.SetDek(record.Dek)
	secret.SetData(record.Data)
	secret.SetVersion(record.Version)
	secret.SetDeleted(record.Deleted)
	op.SetSecret(&secret)
	return &op
}

func toRecord(secret *pb.Secret) *core.Record {
	var record core.Record
	record.ID = secret.GetId()
	record.CreatedAt = time.Now().UTC()
	record.ModifiedAt = secret.GetModifiedAt().AsTime()
	record.BigData = secret.GetIsBig()
	record.Dek = secret.GetDek()
	record.Data = secret.GetData()
	record.Version = secret.GetVersion()
	switch secret.GetType() {
	case pb.SecretType_LoginPass:
		record.Type = core.LoginPassType
	case pb.SecretType_Text:
		record.Type = core.TextType
	case pb.SecretType_BankCard:
		record.Type = core.BankCardType
	case pb.SecretType_Binary:
		record.Type = core.OtherType
	}
	return &record
}

func getState(ctx context.Context, tx *sql.Tx) (*core.SyncState, error) {
	syncState, err := persistence.TxGetState(ctx, tx, syncTypeName)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
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
		secret.SetType(pb.SecretType_LoginPass)
	case core.TextType:
		secret.SetType(pb.SecretType_Text)
	case core.BankCardType:
		secret.SetType(pb.SecretType_BankCard)
	case core.OtherType:
		secret.SetType(pb.SecretType_Binary)
	}
	return &secret
}

type EmptySyncer struct{}

func NewEmptySyncer() *EmptySyncer {
	return &EmptySyncer{}
}
func (e EmptySyncer) Sync(ctx context.Context, option *SyncOption) error {
	return nil
}
