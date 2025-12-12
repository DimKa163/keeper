package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/crypto"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/DimKa163/keeper/internal/datatool"
)

var (
	ErrFileToBig      = errors.New("file is too big")
	ErrConflictExists = errors.New("conflict exists! solve first")
)

type Version int

const (
	_ Version = iota
	Local
	Remote
)

type (
	LoginPassRequest struct {
		Name  string `json:"name"`
		Login string `json:"login"`
		Pass  string `json:"pass"`
		URL   string `json:"url"`
	}
	BankCardRequest struct {
		Name       string `json:"name"`
		CardNumber string `json:"card_number"`
		Expiry     string `json:"expiry"`
		CVV        string `json:"cvv"`
		HolderName string `json:"holder_name"`
		BankName   string `json:"bank_name"`
		CardType   string `json:"card_type"`
		Currency   string `json:"currency"`
		IsPrimary  bool   `json:"is_primary"`
	}
	TextRequest struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	BinaryRequest struct {
		Path string `json:"path"`
	}
)

type DataManager struct {
	db          *sql.DB
	encoder     core.Encoder
	decoder     core.Decoder
	syncManager Syncer
	fp          *datatool.FileProvider
}

func NewDataService(
	db *sql.DB,
	encoder core.Encoder,
	decoder core.Decoder,
	syncManager Syncer,
	fileProvider *datatool.FileProvider,
) *DataManager {
	return &DataManager{
		db,
		encoder,
		decoder,
		syncManager,
		fileProvider,
	}
}

// Get read secret
func (dm *DataManager) Get(ctx context.Context, id string) (*core.Record, error) {
	record, err := persistence.GetRecordByID(ctx, dm.db, id)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// GetAll get all not deleted and not corrupted secrets
func (dm *DataManager) GetAll(ctx context.Context, limit, offset int32) ([]*core.Record, error) {
	return persistence.GetAllRecord(ctx, dm.db, limit, offset)
}

// ExtractFile read file and decode it for export
func (dm *DataManager) ExtractFile(ctx context.Context, record *core.Record) (*core.Binary, io.ReadCloser, error) {
	fs, err := dm.fp.OpenRead(record.ID, record.Version)
	if err != nil {
		return nil, nil, err
	}
	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		return nil, nil, err
	}
	dek, err := dm.decoder.Decode(record.Dek, masterKey)
	if err != nil {
		return nil, nil, err
	}

	md, err := record.DecodeBinary(dm.decoder, masterKey)
	if err != nil {
		return nil, nil, err
	}
	return md, crypto.NewFileDecoder(dm.decoder, fs, dek), nil
}

// GetAllConflicts get all conflicts
func (dm *DataManager) GetAllConflicts(ctx context.Context) ([]*core.Conflict, error) {
	return persistence.GetAllConflict(ctx, dm.db)
}

// Decode decrypt secret
func (dm *DataManager) Decode(ctx context.Context, record *core.Record) ([]byte, error) {
	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		return nil, err
	}
	return record.Decode(dm.decoder, masterKey)
}

// SolveConflict solve conflict between client version and server version
func (dm *DataManager) SolveConflict(ctx context.Context, version Version, conflict *core.Conflict) error {
	tx, err := dm.db.BeginTx(ctx, nil)
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
	switch version {
	case Local:
		if err = dm.applyLocal(ctx, tx, conflict); err != nil {
			return err
		}
	case Remote:
		if err = dm.applyRemote(ctx, tx, conflict); err != nil {
			return err
		}
	}
	return nil
}

// DeleteConflict delete conflict
func (dm *DataManager) DeleteConflict(ctx context.Context, conflict *core.Conflict) error {
	return persistence.DeleteConflict(ctx, dm.db, conflict.ID)
}

func (dm *DataManager) CreateLoginPass(ctx context.Context, request *LoginPassRequest, sync bool) (string, error) {
	var err error
	var id string
	id, err = dm.execInsert(ctx, func(ctx context.Context, tx *sql.Tx) (*core.Record, error) {
		return dm.processLoginPass(
			ctx,
			core.CreateRecord(core.LoginPassType),
			request,
		)
	})
	if sync {
		if err = dm.syncManager.Sync(ctx, &SyncOption{}); err != nil {
			return "", err
		}
	}
	return id, nil
}

func (dm *DataManager) UpdateLoginPass(ctx context.Context, id string, request *LoginPassRequest, sync bool) (string, error) {
	var err error
	id, err = dm.execUpdate(ctx, id, func(ctx context.Context, tx *sql.Tx, record *core.Record) (*core.Record, error) {
		return dm.processLoginPass(
			ctx,
			record,
			request,
		)
	})
	if sync {
		if err = dm.syncManager.Sync(ctx, &SyncOption{}); err != nil {
			return "", err
		}
	}
	return id, nil
}

func (dm *DataManager) CreateText(ctx context.Context, request *TextRequest, sync bool) (string, error) {
	var err error
	var id string
	id, err = dm.execInsert(ctx, func(ctx context.Context, tx *sql.Tx) (*core.Record, error) {
		return dm.processText(
			ctx,
			core.CreateRecord(core.TextType),
			request,
		)
	})
	if sync {
		if err = dm.syncManager.Sync(ctx, &SyncOption{}); err != nil {
			return "", err
		}
	}
	return id, nil
}

func (dm *DataManager) UpdateText(ctx context.Context, id string, request *TextRequest, sync bool) (string, error) {
	var err error
	id, err = dm.execUpdate(ctx, id, func(ctx context.Context, tx *sql.Tx, record *core.Record) (*core.Record, error) {
		return dm.processText(
			ctx,
			record,
			request,
		)
	})
	if sync {
		if err = dm.syncManager.Sync(ctx, &SyncOption{}); err != nil {
			return "", err
		}
	}
	return id, nil
}

func (dm *DataManager) CreateBankCard(ctx context.Context, req *BankCardRequest, sync bool) (string, error) {
	var err error
	var id string
	id, err = dm.execInsert(ctx, func(ctx context.Context, tx *sql.Tx) (*core.Record, error) {
		return dm.processBankCard(
			ctx,
			core.CreateRecord(core.BankCardType),
			req,
		)
	})
	if sync {
		if err = dm.syncManager.Sync(ctx, &SyncOption{}); err != nil {
			return "", err
		}
	}
	return id, nil
}

func (dm *DataManager) UpdateBankCard(ctx context.Context, id string, req *BankCardRequest, sync bool) (string, error) {
	var err error
	id, err = dm.execUpdate(ctx, id, func(ctx context.Context, tx *sql.Tx, record *core.Record) (*core.Record, error) {
		return dm.processBankCard(
			ctx,
			record,
			req,
		)
	})
	if sync {
		if err = dm.syncManager.Sync(ctx, &SyncOption{}); err != nil {
			return "", err
		}
	}
	return id, nil
}

func (dm *DataManager) CreateBinary(ctx context.Context, req *BinaryRequest, sync bool) (string, error) {
	var err error
	var id string
	id, err = dm.execInsert(ctx, func(ctx context.Context, tx *sql.Tx) (*core.Record, error) {
		return dm.processBinary(
			ctx,
			core.CreateRecord(core.BankCardType),
			req,
		)
	})
	if sync {
		if err = dm.syncManager.Sync(ctx, &SyncOption{}); err != nil {
			return "", err
		}
	}
	return id, nil
}

func (dm *DataManager) UpdateBinary(ctx context.Context, id string, req *BinaryRequest, sync bool) (string, error) {
	var err error
	id, err = dm.execUpdate(ctx, id, func(ctx context.Context, tx *sql.Tx, record *core.Record) (*core.Record, error) {
		return dm.processBinary(
			ctx,
			record,
			req,
		)
	})
	if sync {
		if err = dm.syncManager.Sync(ctx, &SyncOption{}); err != nil {
			return "", err
		}
	}
	return id, nil
}

func (dm *DataManager) execInsert(ctx context.Context, op func(ctx context.Context, tx *sql.Tx) (*core.Record, error)) (string, error) {
	var err error
	var id string
	tx, err := dm.db.Begin()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()
	ex, err := persistence.TxConflictExist(ctx, tx)
	if err != nil {
		return "", err
	}
	if ex {
		return "", ErrConflictExists
	}
	record, err := op(ctx, tx)
	if err != nil {
		return "", err
	}
	id, err = dm.insert(ctx, tx, record)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (dm *DataManager) execUpdate(ctx context.Context, id string, op func(ctx context.Context, tx *sql.Tx, record *core.Record) (*core.Record, error)) (string, error) {
	var err error
	var record *core.Record
	tx, err := dm.db.Begin()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()
	ex, err := persistence.TxConflictExist(ctx, tx)
	if err != nil {
		return "", err
	}
	if ex {
		return "", ErrConflictExists
	}
	version := common.GetVersion(ctx)
	record, err = persistence.TxGetRecordByID(ctx, tx, id)
	if err != nil {
		return "", err
	}
	record, err = op(ctx, tx, record)
	if err != nil {
		return "", err
	}
	record.Version = version + 1
	record.Corrupted = false
	record, err = dm.update(ctx, tx, record)
	if err != nil {
		return "", err
	}
	return record.ID, nil
}

func (dm *DataManager) Delete(ctx context.Context, id string, sync bool) error {
	if err := dm.deleteRecord(ctx, id); err != nil {
		return err
	}
	if sync {
		if err := dm.syncManager.Sync(ctx, &SyncOption{}); err != nil {
			return err
		}
	}
	return nil
}

// deleteRecord delete secret
func (dm *DataManager) deleteRecord(ctx context.Context, id string) error {
	tx, err := dm.db.BeginTx(ctx, nil)
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
	ex, err := persistence.TxConflictExist(ctx, tx)
	if err != nil {
		return err
	}
	if ex {
		return ErrConflictExists
	}
	version := common.GetVersion(ctx)
	newVersion := version + 1
	record, err := persistence.GetRecordByID(ctx, dm.db, id)
	if err != nil {
		return err
	}
	if record.BigData {
		if err = dm.fp.Rename(record.ID, record.Version, newVersion); err != nil {
			return err
		}
	}
	record.Deleted = true
	record.Version = newVersion
	if _, err = dm.update(ctx, tx, record); err != nil {
		return err
	}
	return nil
}

func (dm *DataManager) processLoginPass(ctx context.Context, record *core.Record, data *LoginPassRequest) (*core.Record, error) {
	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		return nil, err
	}
	version := common.GetVersion(ctx)
	var model core.LoginPass
	if record.Data != nil {
		var it *core.LoginPass
		it, err = record.DecodeLoginPass(dm.decoder, masterKey)
		if err != nil {
			return nil, err
		}
		model = *it
	}
	if data.Name != "" {
		model.Name = data.Name
	}
	if data.Login != "" {
		model.Login = data.Login
	}
	if data.Pass != "" {
		model.Pass = data.Pass
	}
	if data.URL != "" {
		model.URL = data.URL
	}
	js, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	dek, err := datatool.GenerateDek(32)
	if err != nil {
		return nil, err
	}
	cipher, err := dm.encoder.Encode(js, dek)
	if err != nil {
		return nil, err
	}
	record.Data = cipher
	dekCipher, err := dm.encoder.Encode(dek, masterKey)
	if err != nil {
		return nil, err
	}
	record.Dek = dekCipher
	record.Version = version + 1
	return record, nil
}

func (dm *DataManager) processText(ctx context.Context, record *core.Record, data *TextRequest) (*core.Record, error) {
	version := common.GetVersion(ctx)
	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		return nil, err
	}
	var model core.Text
	if record.Data != nil {
		var it *core.Text
		it, err = record.DecodeText(dm.decoder, masterKey)
		if err != nil {
			return nil, err
		}
		model = *it
	}
	if data.Name != "" {
		model.Name = data.Name
	}
	if data.Content != "" {
		model.Content = data.Content
	}
	js, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	dek, err := datatool.GenerateDek(32)
	if err != nil {
		return nil, err
	}
	cipher, err := dm.encoder.Encode(js, dek)
	if err != nil {
		return nil, err
	}
	record.Data = cipher
	dekCipher, err := dm.encoder.Encode(dek, masterKey)
	if err != nil {
		return nil, err
	}
	record.Dek = dekCipher
	record.Version = version + 1
	return record, nil
}

func (dm *DataManager) processBankCard(ctx context.Context, record *core.Record, data *BankCardRequest) (*core.Record, error) {
	version := common.GetVersion(ctx)
	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		return nil, err
	}
	var model core.BankCard
	if record.Data != nil {
		var card *core.BankCard
		card, err = record.DecodeBankCard(dm.decoder, masterKey)
		if err != nil {
			return nil, err
		}
		model = *card
	}
	if data.Name != "" {
		model.Name = data.Name
	}
	if data.CardNumber != "" {
		model.CardNumber = data.CardNumber
	}
	if data.Expiry != "" {
		model.Expiry = data.Expiry
	}
	if data.CVV != "" {
		model.CVV = data.CVV
	}
	if data.HolderName != "" {
		model.HolderName = data.HolderName
	}
	if data.BankName != "" {
		model.BankName = data.BankName
	}
	if data.CardType != "" {
		model.CardType = data.CardType
	}
	if data.Currency != "" {
		model.Currency = data.Currency
	}
	model.IsPrimary = data.IsPrimary
	js, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	dek, err := datatool.GenerateDek(32)
	if err != nil {
		return nil, err
	}
	cipher, err := dm.encoder.Encode(js, dek)
	if err != nil {
		return nil, err
	}
	record.Data = cipher
	dekCipher, err := dm.encoder.Encode(dek, masterKey)
	if err != nil {
		return nil, err
	}
	record.Dek = dekCipher
	record.Version = version + 1
	return record, nil
}

func (dm *DataManager) processBinary(ctx context.Context, record *core.Record, data *BinaryRequest) (*core.Record, error) {
	version := common.GetVersion(ctx)
	stat, err := os.Stat(data.Path)
	if err != nil {
		return nil, err
	}
	if stat.Size() > datatool.MB*50 {
		return nil, ErrFileToBig
	}
	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		return nil, err
	}
	record.BigData = stat.Size() > datatool.MB
	if record.BigData && record.Version > 0 {
		err = dm.fp.Remove(record.ID, record.Version)
		if err != nil {
			return nil, err
		}
	}

	file, err := os.Open(data.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content := make([]byte, stat.Size())
	_, err = file.Read(content)
	if err != nil {
		return nil, err
	}
	model := core.Binary{
		Name:      stat.Name(),
		SizeBytes: stat.Size(),
		MIMEType:  mime.TypeByExtension(filepath.Ext(stat.Name())),
	}
	dek, err := datatool.GenerateDek(32)
	if err != nil {
		return nil, err
	}
	if record.BigData {
		if err = dm.writeFile(content, dek, record, version+1); err != nil {
			return nil, err
		}
	} else {
		model.Content = content
	}
	js, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	mdCipher, err := dm.encoder.Encode(js, dek)
	if err != nil {
		return nil, err
	}
	record.Data = mdCipher
	cipherDek, err := dm.encoder.Encode(dek, masterKey)
	if err != nil {
		return nil, err
	}
	record.Dek = cipherDek
	record.Version = version + 1
	return record, nil
}

func (dm *DataManager) writeFile(content, dek []byte, record *core.Record, version int32) error {
	f, err := dm.fp.OpenWrite(record.ID, version)
	if err != nil {
		return err
	}
	defer f.Close()
	var cipherData []byte
	cipherData, err = dm.encoder.Encode(content, dek)
	if err != nil {
		return err
	}
	if _, err = f.Write(cipherData); err != nil {
		return err
	}
	return nil
}

func (dm *DataManager) insert(ctx context.Context, tx *sql.Tx, record *core.Record) (string, error) {
	date := time.Now().UTC().Truncate(time.Second)
	record.CreatedAt = date
	record.ModifiedAt = date
	if err := persistence.TxInsertRecord(ctx, tx, record); err != nil {
		return "", err
	}
	return record.ID, nil
}

func (dm *DataManager) update(ctx context.Context, tx *sql.Tx, record *core.Record) (*core.Record, error) {
	record.ModifiedAt = time.Now().UTC().Truncate(time.Second)
	if err := persistence.TxUpdateRecord(ctx, tx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func (dm *DataManager) applyLocal(ctx context.Context, tx *sql.Tx, conflict *core.Conflict) error {
	local := conflict.Local.Record
	remote := conflict.Remote.Record
	if remote.BigData {
		if err := dm.fp.Remove(remote.ID, remote.Version, "remote"); err != nil {
			return err
		}
	}
	if err := persistence.TxUpdateRecord(ctx, tx, local); err != nil {
		return err
	}
	return nil
}

func (dm *DataManager) applyRemote(ctx context.Context, tx *sql.Tx, conflict *core.Conflict) error {
	local := conflict.Local.Record
	remote := conflict.Remote.Record
	if local.BigData {
		if err := dm.fp.Remove(local.ID, local.Version); err != nil {
			return err
		}
	}
	if remote.BigData {
		if err := dm.copyFile(remote); err != nil {
			return err
		}
	}
	if err := persistence.TxUpdateRecord(ctx, tx, remote); err != nil {
		return err
	}
	return nil
}

func (dm *DataManager) copyFile(record *core.Record) error {
	var reader io.ReadCloser
	var writer io.WriteCloser
	var err error
	// копирую два файла
	reader, err = dm.fp.OpenRead(record.ID, record.Version, "remote")
	if err != nil {
		return err
	}
	writer, err = dm.fp.OpenWrite(record.ID, record.Version)
	if err != nil {
		return err
	}
	if _, err = io.Copy(writer, reader); err != nil {
		return err
	}

	if err = reader.Close(); err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}

	if err = dm.fp.Remove(record.ID, record.Version, "remote"); err != nil {
		return err
	}
	return nil
}
