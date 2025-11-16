package cli

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/DimKa163/keeper/internal/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"time"
)

var syncTypeName = reflect.TypeOf(core.Record{}).Name()

type SyncService struct {
	client pb.SyncDataClient
	db     *sql.DB
}

func NewSyncService(
	client pb.SyncDataClient,
	db *sql.DB,
) *SyncService {
	return &SyncService{
		client: client,
		db:     db,
	}
}

func (ss *SyncService) Poll(ctx context.Context) error {
	var err error
	var syncState *core.SyncState
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

	syncState, err = persistence.GetState(ctx, tx, syncTypeName)
	if err != nil {
		if !errors.Is(sql.ErrNoRows, err) {
			return err
		}
		syncState = &core.SyncState{
			ID:           syncTypeName,
			LastSyncTime: time.Time{},
		}
	}
	var request pb.PollRequest
	request.SetSince(timestamppb.New(syncState.LastSyncTime))
	response, err := ss.client.Poll(ctx, &request)
	if err != nil {
		return err
	}
	data := response.GetData()
	var record *core.Record
	for _, dt := range data {
		var exist bool
		exist, err = persistence.TxExists(ctx, tx, dt.GetId())
		if err != nil {
			return err
		}
		if exist {
			record, err = persistence.TxGetRecordByID(ctx, tx, dt.GetId())
			if err != nil {
				return err
			}
			if hasConflict(record, dt) {
				if err = persistence.TxInsertConflict(ctx, tx, &core.Conflict{
					Local:  &core.ConflictItem{Record: record},
					Remote: &core.ConflictItem{Record: mapToRecord(dt), Deleted: dt.GetDeleted()},
				}); err != nil {
					return err
				}
				continue
			}
			if dt.GetDeleted() {
				if err = persistence.TxDeleteRecord(ctx, tx, dt.GetId()); err != nil {
					return err
				}
				continue
			}
			record.Version = dt.GetVersion()
			record.Data = dt.GetData()
			record.DataNonce = dt.GetDataNonce()
			record.Dek = dt.GetDek()
			record.DekNonce = dt.GetDekNonce()
			if err = persistence.TxUpdateRecord(ctx, tx, record); err != nil {
				return err
			}
		} else {
			record = mapToRecord(dt)
			if err = persistence.TxInsertRecord(ctx, tx, record); err != nil {
				return err
			}
		}
	}
	syncState.LastSyncTime = time.Now()
	if err = persistence.SaveState(ctx, tx, syncState); err != nil {
		return err
	}
	return nil
}

func mapToRecord(dt *pb.Data) *core.Record {
	var record *core.Record
	switch dt.GetType() {
	case pb.Data_LoginPass:
		record = core.CreateRecord(core.LoginPassType)
	case pb.Data_Text:
		record = core.CreateRecord(core.TextType)
	case pb.Data_BankCard:
		record = core.CreateRecord(core.BankCardType)
	case pb.Data_Other:
		record = core.CreateRecord(core.OtherType)
	}
	record.ID = dt.GetId()
	record.Data = dt.GetData()
	record.DataNonce = dt.GetDataNonce()
	record.Dek = dt.GetDek()
	record.DekNonce = dt.GetDekNonce()
	record.Version = dt.GetVersion()
	return record
}

func hasConflict(record *core.Record, income *pb.Data) bool {
	version := income.GetVersion()
	return version <= record.Version
}
