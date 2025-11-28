package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/DimKa163/keeper/internal/pb"
	"github.com/DimKa163/keeper/internal/shared"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"reflect"
)

var ErrCorruptedData = errors.New("corrupted data")
var syncTypeName = reflect.TypeOf(core.Record{}).Name()

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

func (ss *SyncService) Sync(ctx context.Context) error {
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
	if err = ss.push(ctx, tx); err != nil {
		return err
	}

	if err = ss.poll(ctx, tx); err != nil {
		return err
	}
	return nil
}

func (ss *SyncService) push(ctx context.Context, tx *sql.Tx) error {
	version := GetVersion(ctx)
	records, err := persistence.TxGetAllRecordGreater(ctx, tx, version)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	if err = ss.pushData(ctx, tx, records); err != nil {
		return err
	}

	return nil
}

func (ss *SyncService) pushData(ctx context.Context, tx *sql.Tx, records []*core.Record) error {
	str, err := ss.client.PushStream(ctx)
	if err != nil {
		return err
	}
	for _, record := range records {
		data := mapToData(record)
		if record.BigData {
			if err = ss.pushBigData(ctx, str, record, data); err != nil {
				if errors.Is(err, ErrCorruptedData) {
					if err = persistence.TxCorruptRecord(ctx, tx, true, record.ID); err != nil {
						return err
					}
					continue
				}
				return err
			}
		} else {
			if err = ss.pushDefault(ctx, str, data); err != nil {
				return err
			}
		}
	}

	if _, err = str.CloseAndRecv(); err != nil {
		return err
	}
	return nil
}

func (ss *SyncService) pushBigData(
	ctx context.Context,
	stream grpc.ClientStreamingClient[pb.Push, pb.PushResponse],
	record *core.Record,
	data *pb.Data,
) error {
	file, err := ss.fileProvider.OpenRead(record.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("file corrupted!")
			return ErrCorruptedData
		}
		return err
	}

	var req pb.Push
	req.SetType(pb.RequestType_StartData)
	req.SetData(data)
	if err := stream.Send(&req); err != nil {
		return err
	}

	buffer := make([]byte, shared.MB)
	for {
		var req pb.Push
		n, err := file.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		req.SetData(data)
		req.SetType(pb.RequestType_FilePart)
		req.SetChunk(buffer[:n])
		if err = stream.Send(&req); err != nil {
			return err
		}
	}
	req = pb.Push{}
	req.SetData(data)
	req.SetType(pb.RequestType_EndData)
	if err = stream.Send(&req); err != nil {
		return err
	}
	file.Close()
	return nil
}

func (ss *SyncService) pushDefault(
	ctx context.Context,
	stream grpc.ClientStreamingClient[pb.Push, pb.PushResponse],
	data *pb.Data,
) error {
	var req pb.Push
	req.SetType(pb.RequestType_Default)
	req.SetData(data)
	if err := stream.Send(&req); err != nil {
		return err
	}
	return nil
}

func (ss *SyncService) poll(ctx context.Context, tx *sql.Tx) error {
	var err error
	var syncState *core.SyncState
	syncState, err = getState(ctx, tx)
	if err != nil {
		return err
	}
	var request pb.PollRequest
	request.SetSince(syncState.Value)
	str, err := ss.client.PollStream(ctx, &request)
	if err != nil {
		return err
	}
	for {
		resp, err := str.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		srvVersion := resp.GetServerVersion()
		if syncState.Value != srvVersion {
			syncState.Value = srvVersion
		}
		switch resp.GetType() {
		case pb.RequestType_Default:
			if err := ss.pullDefault(ctx, tx, resp.GetData()); err != nil {
				return err
			}
		case pb.RequestType_StartData:
			if err := ss.pullDefault(ctx, tx, resp.GetData()); err != nil {
				return err
			}
		case pb.RequestType_FilePart:
			data := resp.GetData()
			if err = ss.pullPart(ctx, data.GetId(), data.GetVersion(), resp.GetChunk()); err != nil {
				return err
			}
		case pb.RequestType_EndData:
			data := resp.GetData()
			if err = ss.pullFinish(ctx, tx, data.GetId(), data.GetVersion()); err != nil {
				return err
			}
		}
	}
	if err = persistence.SaveState(ctx, tx, syncState); err != nil {
		return err
	}
	return nil
}

func (ss *SyncService) pullDefault(ctx context.Context, tx *sql.Tx, data *pb.Data) error {
	exist, err := persistence.TxExists(ctx, tx, data.GetId())
	if err != nil {
		return err
	}
	if exist {
		record, err := persistence.TxGetRecordByID(ctx, tx, data.GetId())
		if err != nil {
			return err
		}
		if data.GetDeleted() {
			if err = persistence.TxDeleteRecord(ctx, tx, data.GetId()); err != nil {
				return err
			}
			return nil
		}
		record.BigData = data.GetLarge()
		record.Version = data.GetVersion()
		record.Data = data.GetData()
		record.DataNonce = data.GetDataNonce()
		record.Dek = data.GetDek()
		record.DekNonce = data.GetDekNonce()
		record.Version = data.GetVersion()
		if err = persistence.TxUpdateRecord(ctx, tx, record); err != nil {
			return err
		}
		if record.FilePath != "" {
			if err = ss.fileProvider.Remove(record.FilePath); err != nil {
				return err
			}
		}
	} else {
		record := mapToRecord(data)
		if err = persistence.TxInsertRecord(ctx, tx, record); err != nil {
			return err
		}
	}
	return nil
}

func (ss *SyncService) pullPart(ctx context.Context, id string, version int32, buffer []byte) error {
	file, err := ss.fileProvider.OpenWrite(fmt.Sprintf("%s_%d", id, version))
	if err != nil {
		return err
	}
	defer file.Close()
	written := 0
	for written < len(buffer) {
		n, err := file.Write(buffer[written:])
		if err != nil {
			log.Fatal(err)
		}
		written += n
	}
	return nil
}

func (ss *SyncService) pullFinish(ctx context.Context, tx *sql.Tx, id string, version int32) error {
	record, err := persistence.TxGetRecordByID(ctx, tx, id)
	if err != nil {
		return err
	}
	record.Version = version
	record.FilePath = fmt.Sprintf("%s_%d", id, version)
	if err = persistence.TxUpdateRecord(ctx, tx, record); err != nil {
		return err
	}
	return nil
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

func mapToRecord(dt *pb.Data) *core.Record {
	var record *core.Record
	switch dt.GetType() {
	case pb.Data_LoginPass:
		record = core.CreateRecord(core.LoginPassType, dt.GetVersion())
	case pb.Data_Text:
		record = core.CreateRecord(core.TextType, dt.GetVersion())
	case pb.Data_BankCard:
		record = core.CreateRecord(core.BankCardType, dt.GetVersion())
	case pb.Data_Other:
		record = core.CreateRecord(core.OtherType, dt.GetVersion())
	}
	record.ID = dt.GetId()
	record.Data = dt.GetData()
	record.DataNonce = dt.GetDataNonce()
	record.Dek = dt.GetDek()
	record.DekNonce = dt.GetDekNonce()
	record.Version = dt.GetVersion()
	return record
}

func mapToData(record *core.Record) *pb.Data {
	var data pb.Data
	data.SetId(record.ID)
	data.SetDataNonce(record.DataNonce)
	data.SetData(record.Data)
	data.SetDekNonce(record.DekNonce)
	data.SetDek(record.Dek)
	data.SetVersion(record.Version)
	data.SetFileDataNonce(record.FileNonce)
	data.SetLarge(record.BigData)
	switch record.Type {
	case core.LoginPassType:
		data.SetType(pb.Data_LoginPass)
	case core.TextType:
		data.SetType(pb.Data_Text)
	case core.BankCardType:
		data.SetType(pb.Data_BankCard)
	case core.OtherType:
		data.SetType(pb.Data_Other)
	}
	return &data
}

func hasConflict(record *core.Record, income *pb.Data) bool {
	version := income.GetVersion()
	return version <= record.Version
}
