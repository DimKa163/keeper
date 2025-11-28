package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/crypto"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/DimKa163/keeper/internal/shared"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrFileToBig = errors.New("file is too big")
)

type RecordRequest struct {
	Type       core.DataType
	Path       string `json:"path"`
	Name       string `json:"name"`
	Login      string `json:"login"`
	Pass       string `json:"pass"`
	Url        string `json:"url"`
	Content    string `json:"content"`
	CardNumber string `json:"card_number"`
	Expiry     string `json:"expiry"`
	CVV        string `json:"cvv"`
	HolderName string `json:"holder_name"`
	BankName   string `json:"bank_name,omitempty"`
	CardType   string `json:"card_type,omitempty"`
	Currency   string `json:"currency,omitempty"`
	IsPrimary  bool   `json:"is_primary"`
}

type DataService struct {
	db           *sql.DB
	encoder      core.Encoder
	decoder      core.Decoder
	fileProvider *shared.FileProvider
}

func NewDataService(db *sql.DB, encoder core.Encoder, decoder core.Decoder, fileProvider *shared.FileProvider) *DataService {
	return &DataService{db, encoder, decoder, fileProvider}
}

func (ds *DataService) Get(ctx context.Context, id string) (*core.Record, error) {
	record, err := persistence.GetRecordByID(ctx, ds.db, id)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (ds *DataService) ExtractFile(ctx context.Context, record *core.Record) (*core.Binary, io.ReadCloser, error) {
	fs, err := ds.fileProvider.OpenRead(record.FilePath)
	if err != nil {
		return nil, nil, err
	}
	masterKey, err := GetMasterKey(ctx)
	if err != nil {
		return nil, nil, err
	}
	dek, err := ds.decoder.Decode(record.DekNonce, record.Dek, masterKey)
	if err != nil {
		return nil, nil, err
	}

	md, err := record.DecodeBinary(ds.decoder, masterKey)
	if err != nil {
		return nil, nil, err
	}
	return md, crypto.NewFileDecoder(ds.decoder, fs, dek, record.FileNonce), nil
}

func (ds *DataService) GetAll(ctx context.Context, limit, offset int32) ([]*core.Record, error) {
	return persistence.GetAllRecord(ctx, ds.db, limit, offset)
}

func (ds *DataService) CreateRecord(ctx context.Context, request *RecordRequest) (string, error) {
	var record *core.Record
	var err error
	tx, err := ds.db.Begin()
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
	version := GetVersion(ctx)
	switch request.Type {
	case core.LoginPassType:
		record, err = ds.processLoginPass(
			ctx,
			core.CreateRecord(request.Type, version+1),
			request,
		)
	case core.TextType:
		record, err = ds.processText(
			ctx,
			core.CreateRecord(request.Type, version+1),
			request,
		)
	case core.BankCardType:
		record, err = ds.processBankCard(
			ctx,
			core.CreateRecord(request.Type, version+1),
			request,
		)
	case core.OtherType:
		record, err = ds.processBinary(
			ctx,
			core.CreateRecord(request.Type, version+1),
			request,
		)
	default:
		return "", errors.New("invalid data type")
	}
	if err != nil {
		return "", err
	}
	var id string
	id, err = ds.insert(ctx, tx, record)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (ds *DataService) UpdateRecord(ctx context.Context, id string, request *RecordRequest) (string, error) {
	var err error
	tx, err := ds.db.BeginTx(ctx, nil)
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
	record, err := persistence.TxGetRecordByID(ctx, tx, id)
	if err != nil {
		return "", err
	}
	version := GetVersion(ctx)
	switch record.Type {
	case core.LoginPassType:
		record, err = ds.processLoginPass(
			ctx,
			record,
			request,
		)
	case core.TextType:
		record, err = ds.processText(
			ctx,
			record,
			request,
		)
	case core.BankCardType:
		record, err = ds.processBankCard(
			ctx,
			record,
			request,
		)
	case core.OtherType:
		record, err = ds.processBinary(
			ctx,
			record,
			request,
		)
	default:
		return "", errors.New("invalid data type")
	}
	record.Version = version + 1
	if err != nil {
		return "", err
	}
	record, err = ds.update(ctx, tx, record)
	if err != nil {
		return "", err
	}
	return record.ID, nil
}

func (ds *DataService) DeleteRecord(ctx context.Context, id string) error {
	tx, err := ds.db.BeginTx(ctx, nil)
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
	version := GetVersion(ctx)
	record, err := persistence.GetRecordByID(ctx, ds.db, id)
	if err != nil {
		return err
	}
	record.Deleted = true
	record.Version = version + 1
	if _, err = ds.update(ctx, tx, record); err != nil {
		return err
	}
	return nil
}

func (ds *DataService) processLoginPass(ctx context.Context, record *core.Record, data *RecordRequest) (*core.Record, error) {
	record.ModifiedAt = time.Now()
	masterKey, err := GetMasterKey(ctx)
	if err != nil {
		return nil, err
	}
	var model core.LoginPass
	if record.Data != nil {
		it, err := record.DecodeLoginPass(ds.decoder, masterKey)
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
	if data.Url != "" {
		model.Url = data.Url
	}
	js, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	dek, err := shared.GenerateDek(32)
	if err != nil {
		return nil, err
	}
	nonce, cipher, err := ds.encoder.Encode(js, dek)
	if err != nil {
		return nil, err
	}
	record.DataNonce = nonce
	record.Data = cipher
	dekNonce, dekCipher, err := ds.encoder.Encode(dek, masterKey)
	if err != nil {
		return nil, err
	}
	record.DekNonce = dekNonce
	record.Dek = dekCipher
	return record, nil
}

func (ds *DataService) processText(ctx context.Context, record *core.Record, data *RecordRequest) (*core.Record, error) {
	record.ModifiedAt = time.Now()

	masterKey, err := GetMasterKey(ctx)
	if err != nil {
		return nil, err
	}
	var model core.Text
	if record.Data != nil {
		it, err := record.DecodeText(ds.decoder, masterKey)
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
	dek, err := shared.GenerateDek(32)
	if err != nil {
		return nil, err
	}
	nonce, cipher, err := ds.encoder.Encode(js, dek)
	if err != nil {
		return nil, err
	}
	record.DataNonce = nonce
	record.Data = cipher
	dekNonce, dekCipher, err := ds.encoder.Encode(dek, masterKey)
	if err != nil {
		return nil, err
	}
	record.DekNonce = dekNonce
	record.Dek = dekCipher
	return record, nil
}

func (ds *DataService) processBankCard(ctx context.Context, record *core.Record, data *RecordRequest) (*core.Record, error) {
	record.ModifiedAt = time.Now()
	masterKey, err := GetMasterKey(ctx)
	if err != nil {
		return nil, err
	}
	var model core.BankCard
	if record.Data != nil {
		card, err := record.DecodeBankCard(ds.decoder, masterKey)
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
	if data.IsPrimary {
		model.IsPrimary = true
	}
	js, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	dek, err := shared.GenerateDek(32)
	if err != nil {
		return nil, err
	}
	nonce, cipher, err := ds.encoder.Encode(js, dek)
	if err != nil {
		return nil, err
	}
	record.DataNonce = nonce
	record.Data = cipher
	dekNonce, dekCipher, err := ds.encoder.Encode(dek, masterKey)
	if err != nil {
		return nil, err
	}
	record.DekNonce = dekNonce
	record.Dek = dekCipher
	return record, nil
}

func (ds *DataService) processBinary(ctx context.Context, record *core.Record, data *RecordRequest) (*core.Record, error) {

	stat, err := os.Stat(data.Path)
	if err != nil {
		return nil, err
	}
	if stat.Size() > shared.MB*500 {
		return nil, ErrFileToBig
	}
	masterKey, err := GetMasterKey(ctx)
	if err != nil {
		return nil, err
	}
	record.BigData = stat.Size() > shared.MB*4
	if record.FilePath != "" {
		err = os.Remove(record.FilePath)
		if err != nil {
			return nil, err
		}
	}
	record.FilePath = fmt.Sprintf("%s_%d", record.ID, record.Version)

	file, err := os.Open(data.Path)
	if err != nil {
		return nil, err
	}
	content := make([]byte, stat.Size())
	_, err = file.Read(content)

	model := core.Binary{
		Name:      stat.Name(),
		SizeBytes: stat.Size(),
		MIMEType:  mime.TypeByExtension(filepath.Ext(stat.Name())),
	}
	f, err := ds.fileProvider.OpenWrite(record.FilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dek, err := shared.GenerateDek(32)
	if err != nil {
		return nil, err
	}
	nonce, cipherData, err := ds.encoder.Encode(content, dek)
	if err != nil {
		return nil, err
	}

	record.FileNonce = nonce
	js, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	mdNonce, mdCipher, err := ds.encoder.Encode(js, dek)
	if err != nil {
		return nil, err
	}
	record.DataNonce = mdNonce
	record.Data = mdCipher
	dekNonce, cipherDek, err := ds.encoder.Encode(dek, masterKey)
	if _, err = f.Write(cipherData); err != nil {
		return nil, err
	}
	record.Dek = cipherDek
	record.DekNonce = dekNonce

	record.ModifiedAt = time.Now()
	return record, nil
}

func (ds *DataService) insert(ctx context.Context, tx *sql.Tx, record *core.Record) (string, error) {

	if err := persistence.TxInsertRecord(ctx, tx, record); err != nil {
		return "", err
	}
	return record.ID, nil
}

func (ds *DataService) update(ctx context.Context, tx *sql.Tx, record *core.Record) (*core.Record, error) {
	record.ModifiedAt = time.Now()
	if err := persistence.TxUpdateRecord(ctx, tx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func (ds *DataService) delete(ctx context.Context, tx *sql.Tx, id string) error {
	if err := persistence.TxDeleteRecord(ctx, tx, id); err != nil {
		return err
	}
	return nil
}
