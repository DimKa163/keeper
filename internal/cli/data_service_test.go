package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/DimKa163/keeper/internal/shared"
	"io"
	"os"
	"testing"

	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/crypto"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/stretchr/testify/assert"
)

func TestDataService_CreateLoginPassShouldBeSuccess(t *testing.T) {
	services, db, cli, dbPath := prepareDataServiceTest(t, func(ctx *CLI, services *ServiceContainer) error {
		return nil
	})
	defer os.Remove(dbPath)
	sut := services.DataService
	request := RecordRequest{
		Type:  core.LoginPassType,
		Name:  "test",
		Login: "test",
		Pass:  "test",
		Url:   "https://www.youtube.com/",
	}

	id, err := sut.CreateRecord(cli, &request)

	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	record, err := persistence.GetRecordByID(cli, db, id)
	masterKey, _ := cli.MasterKey()
	assert.NoError(t, err)
	assert.NotNil(t, record)

	d, err := record.DecodeLoginPass(cli.Decoder(), masterKey)
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, request.Name, d.Name)
	assert.Equal(t, request.Login, d.Login)
	assert.Equal(t, request.Pass, d.Pass)
	assert.Equal(t, request.Url, d.Url)
}

func TestDataService_CreateTextShouldBeSuccess(t *testing.T) {
	services, db, cli, dbPath := prepareDataServiceTest(t, func(ctx *CLI, services *ServiceContainer) error {
		return nil
	})
	defer os.Remove(dbPath)
	sut := services.DataService
	request := RecordRequest{
		Type:    core.TextType,
		Name:    "test",
		Content: "Lorem Ipsum",
	}
	id, err := sut.CreateRecord(cli, &request)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	record, err := persistence.GetRecordByID(cli, db, id)
	masterKey, _ := cli.MasterKey()
	assert.NoError(t, err)
	assert.NotNil(t, record)
	d, err := record.DecodeText(cli.Decoder(), masterKey)
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, request.Name, d.Name)
	assert.Equal(t, request.Content, d.Content)
}

func TestDataService_CreateBankCardShouldBeSuccess(t *testing.T) {
	services, db, cli, dbPath := prepareDataServiceTest(t, func(ctx *CLI, services *ServiceContainer) error {
		return nil
	})
	defer os.Remove(dbPath)
	sut := services.DataService
	request := RecordRequest{
		Type:       core.BankCardType,
		Name:       "test",
		CardNumber: "1232 1232 2345 3234",
		Expiry:     "09/30",
		CVV:        "321",
		HolderName: "Danil",
		BankName:   "Leaky Bucket",
		CardType:   "VISA",
		Currency:   "USD",
		IsPrimary:  true,
	}
	id, err := sut.CreateRecord(cli, &request)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	record, err := persistence.GetRecordByID(cli, db, id)
	masterKey, _ := cli.MasterKey()
	assert.NoError(t, err)
	assert.NotNil(t, record)
	d, err := record.DecodeBankCard(cli.Decoder(), masterKey)
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, request.Name, d.Name)
	assert.Equal(t, request.CardNumber, d.CardNumber)
	assert.Equal(t, request.Expiry, d.Expiry)
	assert.Equal(t, request.CVV, d.CVV)
	assert.Equal(t, request.HolderName, d.HolderName)
	assert.Equal(t, request.BankName, d.BankName)
	assert.Equal(t, request.CardType, d.CardType)
	assert.Equal(t, request.IsPrimary, d.IsPrimary)
}

func TestDataService_CreateBinaryShouldBeSuccess(t *testing.T) {
	services, db, cli, dbPath := prepareDataServiceTest(t, func(ctx *CLI, services *ServiceContainer) error {
		return nil
	})
	defer os.Remove(dbPath)
	path := "temp.zip"
	temp, err := os.Create("temp.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	buff := make([]byte, 1024*1024*50)
	_, _ = rand.Read(buff)
	_, err = io.Copy(temp, bytes.NewReader(buff))
	_ = temp.Close()

	sut := services.DataService
	request := RecordRequest{
		Type: core.OtherType,
		Path: path,
	}
	id, err := sut.CreateRecord(cli, &request)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	record, err := persistence.GetRecordByID(cli, db, id)
	masterKey, _ := cli.MasterKey()
	assert.NoError(t, err)
	assert.NotNil(t, record)
	d, err := record.DecodeBinary(cli.Decoder(), masterKey)
	assert.NoError(t, err)
	assert.NotNil(t, d)
}

func TestDataService_UpdateLoginPassShouldBeSuccess(t *testing.T) {
	var recordID string
	services, db, cli, dbPath := prepareDataServiceTest(t, func(ctx *CLI, services *ServiceContainer) error {
		var err error
		recordID, err = services.DataService.CreateRecord(ctx, &RecordRequest{
			Type:  core.LoginPassType,
			Name:  "test",
			Login: "test",
			Pass:  "test",
			Url:   "https://www.youtube.com/",
		})
		if err != nil {
			return err
		}
		return nil
	})
	defer os.Remove(dbPath)
	sut := services.DataService
	request := RecordRequest{
		Type:  core.LoginPassType,
		Name:  "test",
		Login: "test",
		Pass:  "test123",
		Url:   "https://www.youtube.com/",
	}
	rec, err := sut.UpdateRecord(cli, recordID, &request)
	assert.NoError(t, err)
	assert.NotEmpty(t, rec)

	record, err := persistence.GetRecordByID(cli, db, recordID)
	masterKey, _ := cli.MasterKey()
	assert.NoError(t, err)
	assert.NotNil(t, record)

	d, err := record.DecodeLoginPass(cli.Decoder(), masterKey)
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, request.Name, d.Name)
	assert.Equal(t, request.Login, d.Login)
	assert.Equal(t, request.Pass, d.Pass)
	assert.Equal(t, request.Url, d.Url)
}

func TestDataService_UpdateTextShouldBeSuccess(t *testing.T) {
	var recordID string
	services, db, cli, dbPath := prepareDataServiceTest(t, func(ctx *CLI, services *ServiceContainer) error {
		var err error
		recordID, err = services.DataService.CreateRecord(ctx, &RecordRequest{
			Type:    core.TextType,
			Name:    "test",
			Content: "Hello World",
		})
		if err != nil {
			return err
		}
		return nil
	})
	defer os.Remove(dbPath)
	sut := services.DataService
	request := RecordRequest{
		Type:    core.TextType,
		Name:    "test",
		Content: "Hi World",
	}
	rec, err := sut.UpdateRecord(cli, recordID, &request)
	assert.NoError(t, err)
	assert.NotEmpty(t, rec)

	record, err := persistence.GetRecordByID(cli, db, recordID)
	masterKey, _ := cli.MasterKey()
	assert.NoError(t, err)
	assert.NotNil(t, record)

	d, err := record.DecodeText(cli.Decoder(), masterKey)
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, request.Name, d.Name)
	assert.Equal(t, request.Content, d.Content)
}

func TestDataService_UpdateBankCardShouldBeSuccess(t *testing.T) {
	var recordID string
	services, db, cli, dbPath := prepareDataServiceTest(t, func(ctx *CLI, services *ServiceContainer) error {
		var err error
		recordID, err = services.DataService.CreateRecord(ctx, &RecordRequest{
			Type:       core.BankCardType,
			Name:       "test",
			CardNumber: "1232 1232 2345 3234",
			Expiry:     "09/30",
			CVV:        "321",
			HolderName: "Danil",
			BankName:   "Leaky Bucket",
			CardType:   "VISA",
			Currency:   "USD",
			IsPrimary:  true,
		})
		if err != nil {
			return err
		}
		return nil
	})
	defer os.Remove(dbPath)
	sut := services.DataService
	request := RecordRequest{
		Type:       core.BankCardType,
		Name:       "test",
		CardNumber: "1232 1232 2345 3234",
		Expiry:     "09/30",
		CVV:        "321",
		HolderName: "Danil",
		BankName:   "Leaky Bucket",
		CardType:   "VISA",
		Currency:   "USD",
		IsPrimary:  false,
	}
	rec, err := sut.UpdateRecord(cli, recordID, &request)
	assert.NoError(t, err)
	assert.NotEmpty(t, rec)

	record, err := persistence.GetRecordByID(cli, db, recordID)
	masterKey, _ := cli.MasterKey()
	assert.NoError(t, err)
	assert.NotNil(t, record)

	d, err := record.DecodeBankCard(cli.Decoder(), masterKey)
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, request.Name, d.Name)
	assert.Equal(t, request.CardNumber, d.CardNumber)
	assert.Equal(t, request.Expiry, d.Expiry)
	assert.Equal(t, request.CVV, d.CVV)
	assert.Equal(t, request.HolderName, d.HolderName)
	assert.Equal(t, request.BankName, d.BankName)
	assert.Equal(t, request.CardType, d.CardType)
	assert.Equal(t, request.IsPrimary, d.IsPrimary)
}

func TestDataService_UpdateBinaryShouldBeSuccess(t *testing.T) {
	path := "temp.zip"
	temp1, err := os.Create("temp.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	buff1 := make([]byte, 1024*1024*50)
	_, _ = rand.Read(buff1)
	_, err = io.Copy(temp1, bytes.NewReader(buff1))
	_ = temp1.Close()
	var recordID string
	services, db, cli, dbPath := prepareDataServiceTest(t, func(ctx *CLI, services *ServiceContainer) error {
		var err error
		recordID, err = services.DataService.CreateRecord(ctx, &RecordRequest{
			Type: core.OtherType,
			Path: path,
		})
		if err != nil {
			return err
		}
		return nil
	})
	defer os.Remove(dbPath)
	path = "temp1.zip"
	temp, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	buff := make([]byte, 1024*1024*45)
	_, _ = rand.Read(buff)
	_, err = io.Copy(temp, bytes.NewReader(buff))
	_ = temp1.Close()
	sut := services.DataService
	request := RecordRequest{
		Type: core.OtherType,
		Path: path,
	}
	rec, err := sut.UpdateRecord(cli, recordID, &request)
	assert.NoError(t, err)
	assert.NotEmpty(t, rec)

	record, err := persistence.GetRecordByID(cli, db, recordID)
	masterKey, _ := cli.MasterKey()
	assert.NoError(t, err)
	assert.NotNil(t, record)
	d, err := record.DecodeBinary(cli.Decoder(), masterKey)
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.NotEqual(t, d.Content, buff1)
}

func prepareDataServiceTest(t *testing.T, fn func(cli *CLI, services *ServiceContainer) error) (*ServiceContainer, *sql.DB, *CLI, string) {
	path, err := createDirIfNotExist()
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s", path))
	if err != nil {
		t.Fatal(err)
	}
	if err = persistence.Migrate(db, "./migrations"); err != nil {
		t.Fatal(err)
	}
	//recordRepository := persistence.NewRecordRepository(db)
	//userRepository := persistence.NewUserRepository(db)
	services := &ServiceContainer{
		Decoder:     crypto.NewAesDecoder(),
		Encoder:     crypto.NewAesEncoder(),
		DataService: NewDataService(db),
	}
	masterKey := "123qweASD"
	salt, err := shared.GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	hash := shared.Hash([]byte(masterKey), salt, 2, 64, 32, 2)
	cli := NewCLI(context.Background(), services, hash)
	if err := fn(cli, services); err != nil {
		t.Fatal(err)
	}
	return services, db, cli, path
}

func createDirIfNotExist() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := homeDir + "\\.keeper"
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s\\data_service_text_keeper.db", dir), err
}
