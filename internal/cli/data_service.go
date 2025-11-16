package cli

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/persistence"
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
	db *sql.DB
}

func NewDataService(db *sql.DB) *DataService {
	return &DataService{db}
}

func (ds *DataService) CreateRecord(ctx *CLI, request *RecordRequest) (string, error) {
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
	switch request.Type {
	case core.LoginPassType:
		record, err = ds.processLoginPass(
			ctx,
			core.CreateRecord(request.Type),
			request,
		)
	case core.TextType:
		record, err = ds.processText(
			ctx,
			core.CreateRecord(request.Type),
			request,
		)
	case core.BankCardType:
		record, err = ds.processBankCard(
			ctx,
			core.CreateRecord(request.Type),
			request,
		)
	case core.OtherType:
		record, err = ds.processBinary(
			ctx,
			core.CreateRecord(request.Type),
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

func (ds *DataService) UpdateRecord(ctx *CLI, id string, request *RecordRequest) (*core.Record, error) {
	var err error
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
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
		return nil, err
	}
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
		return nil, errors.New("invalid data type")
	}
	if err != nil {
		return nil, err
	}
	record, err = ds.update(ctx, tx, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (ds *DataService) DeleteRecord(ctx *CLI, id string) error {
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
	record, err := persistence.GetRecordByID(ctx, ds.db, id)
	if err != nil {
		return err
	}
	return ds.delete(ctx, tx, record.ID)
}

func (ds *DataService) processLoginPass(ctx *CLI, record *core.Record, data *RecordRequest) (*core.Record, error) {
	record.ModifiedAt = time.Now()
	model := &core.LoginPass{
		Name:  data.Name,
		Login: data.Login,
		Pass:  data.Pass,
		Url:   data.Url,
	}
	js, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	masterKey, ok := ctx.MasterKey()
	if !ok {
		return nil, errors.New("master key not set")
	}
	if err = record.Encode(ctx.Encoder(), js, masterKey); err != nil {
		return nil, err
	}
	return record, nil
}

func (ds *DataService) processText(ctx *CLI, record *core.Record, data *RecordRequest) (*core.Record, error) {
	record.ModifiedAt = time.Now()
	model := &core.Text{
		Name:    data.Name,
		Content: data.Content,
	}
	js, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	masterKey, ok := ctx.MasterKey()
	if !ok {
		return nil, errors.New("master key not set")
	}
	if err = record.Encode(ctx.Encoder(), js, masterKey); err != nil {
		return nil, err
	}
	return record, nil
}

func (ds *DataService) processBankCard(ctx *CLI, record *core.Record, data *RecordRequest) (*core.Record, error) {
	record.ModifiedAt = time.Now()
	model := &core.BankCard{
		Name:       data.Name,
		CardNumber: data.CardNumber,
		Expiry:     data.Expiry,
		CVV:        data.CVV,
		HolderName: data.HolderName,
		BankName:   data.BankName,
		CardType:   data.CardType,
		Currency:   data.Currency,
		IsPrimary:  data.IsPrimary,
	}
	js, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	masterKey, ok := ctx.MasterKey()
	if !ok {
		return nil, errors.New("master key not set")
	}
	if err = record.Encode(ctx.Encoder(), js, masterKey); err != nil {
		return nil, err
	}
	return record, nil
}

func (ds *DataService) processBinary(ctx *CLI, record *core.Record, data *RecordRequest) (*core.Record, error) {
	record.ModifiedAt = time.Now()
	stat, err := os.Stat(data.Path)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(data.Path)
	if err != nil {
		return nil, err
	}
	content := make([]byte, stat.Size())
	_, err = file.Read(content)
	if err != nil {
		return nil, err
	}
	model := &core.Binary{
		Name:      stat.Name(),
		SizeBytes: stat.Size(),
		Content:   content,
	}
	js, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	masterKey, ok := ctx.MasterKey()
	if !ok {
		return nil, errors.New("master key not set")
	}
	if err = record.Encode(ctx.Encoder(), js, masterKey); err != nil {
		return nil, err
	}
	return record, nil
}

func (ds *DataService) insert(ctx *CLI, tx *sql.Tx, record *core.Record) (string, error) {

	if err := persistence.TxInsertRecord(ctx, tx, record); err != nil {
		return "", err
	}
	return record.ID, nil
}

func (ds *DataService) update(ctx *CLI, tx *sql.Tx, record *core.Record) (*core.Record, error) {
	record.ModifiedAt = time.Now()
	if err := persistence.TxUpdateRecord(ctx, tx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func (ds *DataService) delete(ctx *CLI, tx *sql.Tx, id string) error {
	if err := persistence.TxDeleteRecord(ctx, tx, id); err != nil {
		return err
	}
	return nil
}
