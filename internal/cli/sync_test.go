package cli

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/crypto"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/DimKa163/keeper/internal/mocks"
	"github.com/DimKa163/keeper/internal/pb"
	"github.com/DimKa163/keeper/internal/shared"
	"github.com/beevik/guid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSyncService_PollShouldCreateRecordSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockClient := mocks.NewMockSyncDataClient(ctrl)
	servs, db, _, path := prepareSyncService(t, mockClient, func(cli *CLI, services *ServiceContainer) error {
		return nil
	})
	defer os.Remove(path)
	var resp pb.PollResponse
	id := guid.NewString()
	var data pb.Data
	dek, dekNonce := make([]byte, 32), make([]byte, 16)
	_, _ = rand.Read(dek)
	_, _ = rand.Read(dekNonce)
	dt, dtNonce := make([]byte, 4026), make([]byte, 16)
	_, _ = rand.Read(dt)
	_, _ = rand.Read(dtNonce)
	data.SetName("login/pass")
	data.SetData(dt)
	data.SetDataNonce(dtNonce)
	data.SetDek(dek)
	data.SetDekNonce(dekNonce)
	data.SetLarge(false)
	data.SetType(pb.Data_LoginPass)
	data.SetVersion(1)
	data.SetId(id)
	dataSl := make([]*pb.Data, 0)
	dataSl = append(dataSl, &data)
	resp.SetData(dataSl)
	mockClient.EXPECT().Poll(ctx, gomock.Any()).Return(&resp, nil)
	sut := servs.SyncService

	err := sut.Poll(ctx)

	assert.NoError(t, err)

	r, err := persistence.GetRecordByID(ctx, db, id)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}

func TestSyncService_PollShouldUpdateRecordSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockClient := mocks.NewMockSyncDataClient(ctrl)
	id := guid.NewString()
	servs, db, _, path := prepareSyncService(t, mockClient, func(cli *CLI, services *ServiceContainer) error {
		dek, dekNonce := make([]byte, 32), make([]byte, 16)
		_, _ = rand.Read(dek)
		_, _ = rand.Read(dekNonce)
		dt, dtNonce := make([]byte, 4026), make([]byte, 16)
		_, _ = rand.Read(dt)
		_, _ = rand.Read(dtNonce)
		var record core.Record
		record.ID = id
		record.Type = core.LoginPassType
		record.Data = dt
		record.DataNonce = dtNonce
		record.Dek = dek
		record.DekNonce = dekNonce
		record.Version = 1
		return persistence.InsertRecord(cli, services.DB, &record)
	})
	defer os.Remove(path)
	var resp pb.PollResponse

	var data pb.Data
	dek, dekNonce := make([]byte, 32), make([]byte, 16)
	_, _ = rand.Read(dek)
	_, _ = rand.Read(dekNonce)
	dt, dtNonce := make([]byte, 4026), make([]byte, 16)
	_, _ = rand.Read(dt)
	_, _ = rand.Read(dtNonce)
	data.SetName("login/pass")
	data.SetData(dt)
	data.SetDataNonce(dtNonce)
	data.SetDek(dek)
	data.SetDekNonce(dekNonce)
	data.SetLarge(false)
	data.SetType(pb.Data_LoginPass)
	data.SetVersion(2)
	data.SetId(id)
	dataSl := make([]*pb.Data, 0)
	dataSl = append(dataSl, &data)
	resp.SetData(dataSl)
	mockClient.EXPECT().Poll(ctx, gomock.Any()).Return(&resp, nil)
	sut := servs.SyncService

	err := sut.Poll(ctx)

	assert.NoError(t, err)

	r, err := persistence.GetRecordByID(ctx, db, id)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, data.GetDataNonce(), r.DataNonce)
	assert.Equal(t, data.GetData(), r.Data)
	assert.Equal(t, data.GetVersion(), r.Version)
}

func TestSyncService_PollShouldDeleteRecordSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockClient := mocks.NewMockSyncDataClient(ctrl)
	id := guid.NewString()
	servs, db, _, path := prepareSyncService(t, mockClient, func(cli *CLI, services *ServiceContainer) error {
		dek, dekNonce := make([]byte, 32), make([]byte, 16)
		_, _ = rand.Read(dek)
		_, _ = rand.Read(dekNonce)
		dt, dtNonce := make([]byte, 4026), make([]byte, 16)
		_, _ = rand.Read(dt)
		_, _ = rand.Read(dtNonce)
		var record core.Record
		record.ID = id
		record.Type = core.LoginPassType
		record.Data = dt
		record.DataNonce = dtNonce
		record.Dek = dek
		record.DekNonce = dekNonce
		record.Version = 1
		return persistence.InsertRecord(cli, services.DB, &record)
	})
	defer os.Remove(path)
	var resp pb.PollResponse

	var data pb.Data
	dek, dekNonce := make([]byte, 32), make([]byte, 16)
	_, _ = rand.Read(dek)
	_, _ = rand.Read(dekNonce)
	dt, dtNonce := make([]byte, 4026), make([]byte, 16)
	_, _ = rand.Read(dt)
	_, _ = rand.Read(dtNonce)
	data.SetName("login/pass")
	data.SetData(dt)
	data.SetDataNonce(dtNonce)
	data.SetDek(dek)
	data.SetDekNonce(dekNonce)
	data.SetLarge(false)
	data.SetType(pb.Data_LoginPass)
	data.SetVersion(2)
	data.SetDeleted(true)
	data.SetId(id)
	dataSl := make([]*pb.Data, 0)
	dataSl = append(dataSl, &data)
	resp.SetData(dataSl)
	mockClient.EXPECT().Poll(ctx, gomock.Any()).Return(&resp, nil)
	sut := servs.SyncService

	err := sut.Poll(ctx)

	assert.NoError(t, err)

	r, err := persistence.GetRecordByID(ctx, db, id)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, r)
}

func TestSyncService_PollShouldConflictRecordSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockClient := mocks.NewMockSyncDataClient(ctrl)
	id := guid.NewString()
	servs, db, _, path := prepareSyncService(t, mockClient, func(cli *CLI, services *ServiceContainer) error {
		dek, dekNonce := make([]byte, 32), make([]byte, 16)
		_, _ = rand.Read(dek)
		_, _ = rand.Read(dekNonce)
		dt, dtNonce := make([]byte, 4026), make([]byte, 16)
		_, _ = rand.Read(dt)
		_, _ = rand.Read(dtNonce)
		var record core.Record
		record.ID = id
		record.Type = core.LoginPassType
		record.Data = dt
		record.DataNonce = dtNonce
		record.Dek = dek
		record.DekNonce = dekNonce
		record.Version = 2
		return persistence.InsertRecord(cli, services.DB, &record)
	})
	defer os.Remove(path)
	var resp pb.PollResponse

	var data pb.Data
	dek, dekNonce := make([]byte, 32), make([]byte, 16)
	_, _ = rand.Read(dek)
	_, _ = rand.Read(dekNonce)
	dt, dtNonce := make([]byte, 4026), make([]byte, 16)
	_, _ = rand.Read(dt)
	_, _ = rand.Read(dtNonce)
	data.SetName("login/pass")
	data.SetData(dt)
	data.SetDataNonce(dtNonce)
	data.SetDek(dek)
	data.SetDekNonce(dekNonce)
	data.SetLarge(false)
	data.SetType(pb.Data_LoginPass)
	data.SetVersion(2)
	data.SetDeleted(true)
	data.SetId(id)
	dataSl := make([]*pb.Data, 0)
	dataSl = append(dataSl, &data)
	resp.SetData(dataSl)
	mockClient.EXPECT().Poll(ctx, gomock.Any()).Return(&resp, nil)
	sut := servs.SyncService

	err := sut.Poll(ctx)

	assert.NoError(t, err)

	r, err := persistence.GetRecordByID(ctx, db, id)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	c, err := persistence.ConflictExist(ctx, db)
	assert.NoError(t, err)
	assert.True(t, c)
}

func prepareSyncService(t *testing.T, client *mocks.MockSyncDataClient, fn func(cli *CLI, services *ServiceContainer) error) (*ServiceContainer, *sql.DB, *CLI, string) {
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
	services := &ServiceContainer{
		DB:          db,
		Decoder:     crypto.NewAesDecoder(),
		Encoder:     crypto.NewAesEncoder(),
		SyncService: NewSyncService(client, db),
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
