package app

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/crypto"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/DimKa163/keeper/internal/shared"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
	"os"
	"testing"
)

func TestCreateLoginPassShouldBeSuccess(t *testing.T) {
	ctx, manager, cleanUp := configure(t)

	request := &RecordRequest{
		Type: core.LoginPassType, Name: "Test", Login: "Login", Pass: "Pass", Url: "http:",
	}
	id, err := manager.Create(ctx, request, false)

	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	r, err := persistence.GetRecordByID(ctx, manager.db, id)
	if err != nil {
		t.Fatal(err)
	}

	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	lp, err := r.DecodeLoginPass(manager.decoder, masterKey)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, lp)
	assert.Equal(t, int32(1), r.Version)
	assert.Equal(t, "Test", lp.Name)
	assert.Equal(t, "Login", lp.Login)
	assert.Equal(t, "Pass", lp.Pass)
	assert.Equal(t, "http:", lp.Url)

	if err := manager.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cleanUp(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateLoginPassShouldBeSuccess(t *testing.T) {
	ctx, manager, cleanUp := configure(t)
	id, err := createLoginPass(ctx, manager)
	if err != nil {
		t.Fatal(err)
	}
	request := &RecordRequest{
		Type: core.LoginPassType, Name: "NoTest", Login: "NoLogin", Pass: "NoPass", Url: "http:",
	}
	id, err = manager.Update(ctx, id, request, false)

	r, err := persistence.GetRecordByID(ctx, manager.db, id)
	if err != nil {
		t.Fatal(err)
	}

	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	lp, err := r.DecodeLoginPass(manager.decoder, masterKey)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, int32(1), r.Version)
	assert.Equal(t, request.Name, lp.Name)
	assert.Equal(t, request.Login, lp.Login)
	assert.Equal(t, request.Url, lp.Url)
	assert.Equal(t, request.Pass, lp.Pass)
	if err := manager.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cleanUp(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateTextContentShouldBeSuccess(t *testing.T) {
	ctx, manager, cleanUp := configure(t)

	request := &RecordRequest{
		Type: core.TextType, Name: "test text content", Content: "yep, its content",
	}
	id, err := manager.Create(ctx, request, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	r, err := persistence.GetRecordByID(ctx, manager.db, id)
	if err != nil {
		t.Fatal(err)
	}

	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	txt, err := r.DecodeText(manager.decoder, masterKey)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, txt)
	assert.Equal(t, int32(1), r.Version)
	assert.Equal(t, request.Name, txt.Name)
	assert.Equal(t, request.Content, txt.Content)
	if err := manager.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cleanUp(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateTextContentShouldBeSuccess(t *testing.T) {
	ctx, manager, cleanUp := configure(t)
	id, err := createTextContent(ctx, manager)
	if err != nil {
		t.Fatal(err)
	}
	request := &RecordRequest{
		Type: core.TextType, Name: "test text content, updated", Content: "yep, its content",
	}
	id, err = manager.Update(ctx, id, request, false)

	r, err := persistence.GetRecordByID(ctx, manager.db, id)
	if err != nil {
		t.Fatal(err)
	}

	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	lp, err := r.DecodeText(manager.decoder, masterKey)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, int32(1), r.Version)
	assert.Equal(t, request.Name, lp.Name)
	assert.Equal(t, request.Content, lp.Content)
	if err := manager.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cleanUp(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateBankCardShouldBeSuccess(t *testing.T) {
	ctx, manager, cleanUp := configure(t)

	request := &RecordRequest{
		Type:       core.BankCardType,
		Name:       "test bank card",
		CardNumber: "4001 4547 4778 9852",
		Expiry:     "30/01",
		CVV:        "123",
		HolderName: "IVAN",
		BankName:   "TINKOFF",
		CardType:   "VISA",
		Currency:   "USD",
		IsPrimary:  true,
	}
	id, err := manager.Create(ctx, request, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	r, err := persistence.GetRecordByID(ctx, manager.db, id)
	if err != nil {
		t.Fatal(err)
	}

	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	bc, err := r.DecodeBankCard(manager.decoder, masterKey)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, bc)
	assert.Equal(t, int32(1), r.Version)
	assert.Equal(t, request.Name, bc.Name)
	assert.Equal(t, request.CardType, bc.CardType)
	assert.Equal(t, request.CardNumber, bc.CardNumber)
	assert.Equal(t, request.Expiry, bc.Expiry)
	assert.Equal(t, request.CVV, bc.CVV)
	assert.Equal(t, request.Currency, bc.Currency)
	assert.Equal(t, request.HolderName, bc.HolderName)
	assert.Equal(t, request.BankName, bc.BankName)
	assert.Equal(t, request.IsPrimary, bc.IsPrimary)
	if err := manager.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cleanUp(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateBankCardShouldBeSuccess(t *testing.T) {
	ctx, manager, cleanUp := configure(t)

	id, err := createBankCard(ctx, manager)
	if err != nil {
		t.Fatal(err)
	}
	request := &RecordRequest{
		Type:       core.BankCardType,
		Name:       "test bank card",
		CardNumber: "4001 4547 4778 9852",
		Expiry:     "30/01",
		CVV:        "123",
		HolderName: "IVAN",
		BankName:   "TINKOFF",
		CardType:   "VISA",
		Currency:   "USD",
		IsPrimary:  false,
	}
	id, err = manager.Update(ctx, id, request, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	r, err := persistence.GetRecordByID(ctx, manager.db, id)
	if err != nil {
		t.Fatal(err)
	}
	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	bc, err := r.DecodeBankCard(manager.decoder, masterKey)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, bc)
	assert.Equal(t, int32(1), r.Version)
	assert.Equal(t, request.Name, bc.Name)
	assert.Equal(t, request.CardType, bc.CardType)
	assert.Equal(t, request.CardNumber, bc.CardNumber)
	assert.Equal(t, request.Expiry, bc.Expiry)
	assert.Equal(t, request.CVV, bc.CVV)
	assert.Equal(t, request.Currency, bc.Currency)
	assert.Equal(t, request.HolderName, bc.HolderName)
	assert.Equal(t, request.BankName, bc.BankName)
	assert.Equal(t, request.IsPrimary, bc.IsPrimary)
	if err := manager.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cleanUp(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateBigBinaryShouldBeSuccess(t *testing.T) {
	ctx, manager, cleanUp := configure(t)
	filePath := fmt.Sprintf("%s\\%s", manager.fp.Path, "binary.bin")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	content := make([]byte, shared.MB*5)
	_, err = rand.Read(content)
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.Write(content)
	if err != nil {
		t.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}

	request := &RecordRequest{
		Type: core.OtherType,
		Path: filePath,
	}
	id, err := manager.Create(ctx, request, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	r, err := persistence.GetRecordByID(ctx, manager.db, id)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, r)
	if err := manager.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cleanUp(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateNotBigBinaryShouldBeSuccess(t *testing.T) {
	ctx, manager, cleanUp := configure(t)
	filePath := fmt.Sprintf("%s\\%s", manager.fp.Path, "binary.bin")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	content := make([]byte, shared.MB-2)
	_, err = rand.Read(content)
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.Write(content)
	if err != nil {
		t.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}

	request := &RecordRequest{
		Type: core.OtherType,
		Path: filePath,
	}
	id, err := manager.Create(ctx, request, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	r, err := persistence.GetRecordByID(ctx, manager.db, id)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, r)
	if err := manager.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cleanUp(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateBinaryShouldBeSuccess(t *testing.T) {
	ctx, manager, cleanUp := configure(t)
	filePath := fmt.Sprintf("%s\\%s", manager.fp.Path, "TestUpdateBinaryShouldBeSuccess_old.bin")

	id, err := createBinaryFile(ctx, filePath, manager)
	filePath = fmt.Sprintf("%s\\%s", manager.fp.Path, "TestUpdateBinaryShouldBeSuccess_new.bin")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	content := make([]byte, shared.MB*7)
	_, err = rand.Read(content)
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.Write(content)
	if err != nil {
		t.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}

	request := &RecordRequest{
		Type: core.OtherType,
		Path: filePath,
	}
	id, err = manager.Update(ctx, id, request, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	r, err := persistence.GetRecordByID(ctx, manager.db, id)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, r)

	if err := manager.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cleanUp(); err != nil {
		t.Fatal(err)
	}
}

func configure(t *testing.T) (context.Context, *DataManager, func() error) {
	masterKey := make([]byte, 32)

	if _, err := rand.Read(masterKey); err != nil {
		t.Fatal(err)
	}

	ctx := common.SetVersion(common.SetMasterKey(context.Background(), masterKey), 0)

	encoder := crypto.NewAesEncoder()

	decoder := crypto.NewAesDecoder()
	if err := os.Mkdir("test", os.ModePerm); err != nil {

	}
	path := "./test"

	fileProvider := shared.NewFileProvider(path)

	db, err := sql.Open("sqlite", "file:memdb1?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	if err := persistence.Migrate(db); err != nil {
		t.Fatal(err)
	}
	return ctx, NewDataService(db, encoder, decoder, &mockSyncer{}, fileProvider), func() error {
		return os.RemoveAll(path)
	}
}
func createLoginPass(ctx context.Context, dataService *DataManager) (string, error) {
	request := &RecordRequest{
		Type: core.LoginPassType, Name: "Test", Login: "Login", Pass: "Pass", Url: "http:",
	}
	return dataService.createRecord(ctx, request)
}

func createTextContent(ctx context.Context, dataService *DataManager) (string, error) {
	request := &RecordRequest{
		Type: core.TextType, Name: "test text content", Content: "yep, its content",
	}
	return dataService.createRecord(ctx, request)
}

func createBankCard(ctx context.Context, dataService *DataManager) (string, error) {
	request := &RecordRequest{
		Type:       core.BankCardType,
		Name:       "test bank card",
		CardNumber: "4001 4547 4778 9852",
		Expiry:     "30/01",
		CVV:        "123",
		HolderName: "IVAN",
		BankName:   "TINKOFF",
		CardType:   "VISA",
		Currency:   "USD",
		IsPrimary:  true,
	}
	return dataService.createRecord(ctx, request)
}

func createBinaryFile(ctx context.Context, filePath string, dataService *DataManager) (string, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	content := make([]byte, shared.MB*5)
	_, err = rand.Read(content)
	if err != nil {
		return "", err
	}
	_, err = file.Write(content)
	if err != nil {
		return "", err
	}
	err = file.Close()
	if err != nil {
		return "", err
	}
	request := &RecordRequest{
		Type: core.OtherType,
		Path: filePath,
	}
	return dataService.createRecord(ctx, request)
}

type mockSyncer struct {
}

func (s *mockSyncer) Sync(ctx context.Context, option *SyncOption) error {
	return nil
}
