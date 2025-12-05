package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/crypto"
	"github.com/DimKa163/keeper/internal/cli/interface"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/DimKa163/keeper/internal/shared"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
	"os"
	"reflect"
	"time"
)

type ServiceContainer struct {
	DB          *sql.DB
	UserService *app.UserService
	SyncService *app.SyncService
	DataService *app.DataManager
	Decoder     core.Decoder
	Encoder     core.Encoder
}

type CMD struct {
	*ServiceContainer
	root *cobra.Command
}

func New() (*CMD, error) {
	dir, path, err := createDirIfNotExist()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s", path))
	if err != nil {
		return nil, err
	}
	if err = persistence.Migrate(db); err != nil {
		return nil, err
	}
	fileProvider := shared.NewFileProvider(fmt.Sprintf("%s\\", dir))
	syncService, err := createSyncService(db, fileProvider)
	if err != nil && errors.Is(app.ErrServerUnavailable, err) {
		fmt.Println("remote server unavailable")
		err = nil
	}
	if err != nil {
		return nil, err
	}
	encoder := crypto.NewGzipEncoder(crypto.NewAesEncoder())
	decoder := crypto.NewGzipDecoder(crypto.NewAesDecoder())
	cmd := &CMD{
		ServiceContainer: &ServiceContainer{
			DB:          db,
			UserService: app.NewUserService(db),
			DataService: app.NewDataService(db, encoder, decoder, syncService, fileProvider),
			SyncService: syncService,
			Encoder:     encoder,
			Decoder:     decoder,
		},
	}

	rootCmd := &cobra.Command{
		Use:  "keeper",
		RunE: cmd.Run,
	}
	cmd.root = rootCmd
	return cmd, nil
}

func (cmd *CMD) RegisterCommands() error {
	if err := cmd.addCommand(_interface.NewRegisterBuilder(cmd.UserService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewDataListBuilder(cmd.UserService, cmd.DataService, cmd.Decoder)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewCreateLoginPassBuilder(cmd.UserService, cmd.DataService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewUpdateLoginPassCommandBuilder(cmd.UserService, cmd.DataService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewCreateTextContentBuilder(cmd.UserService, cmd.DataService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewUpdateTextContentCommandBuilder(cmd.UserService, cmd.DataService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewCreateBinaryFileCommandBuilder(cmd.UserService, cmd.DataService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewUpdateBinaryFileCommandBuilder(cmd.UserService, cmd.DataService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewCreateBankCardCommandBuilder(cmd.UserService, cmd.DataService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewUpdateBankCardCommandBuilder(cmd.UserService, cmd.DataService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewDeleteRecordCommandBuilder(cmd.UserService, cmd.DataService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewRegisterRemoteServerCommandBuilder(cmd.UserService, cmd.DB)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewSyncCommandBuilder(cmd.UserService, cmd.SyncService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewReadBinaryFileCommandBuilder(cmd.UserService, cmd.DataService)); err != nil {
		return err
	}
	if err := cmd.addCommand(_interface.NewResolveCommandBuilder(cmd.DataService, cmd.UserService, cmd.SyncService)); err != nil {
	}
	return nil
}

func (cmd *CMD) addCommand(builder core.CommandBuilder) error {
	command, err := builder.Build()
	if err != nil {
		return err
	}
	cmd.root.AddCommand(command)
	return nil
}

func (cmd *CMD) Run(command *cobra.Command, args []string) error {
	fmt.Println(args)
	return nil
}

func (cmd *CMD) Execute() error {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	us, err := persistence.GetUser(ctx, cmd.DB, hostname)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if us != nil {
		ctx = common.SetMasterKey(common.SetHostName(ctx, us.Username), us.Password)
	}
	name := reflect.TypeOf(core.Record{}).Name()
	version, err := persistence.GetState(ctx, cmd.DB, name)
	if err != nil {
		return err
	}
	ctx = common.SetVersion(ctx, version.Value)
	return cmd.root.ExecuteContext(ctx)
}

func createSyncService(db *sql.DB, fp *shared.FileProvider) (*app.SyncService, error) {
	serv, err := persistence.GetServer(context.Background(), db, true)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, nil
	}
	client, err := app.NewRemoteClient(serv.Address, serv.Login, serv.Password)
	if err != nil {
		return nil, err
	}
	if err = client.IsHealthy(context.Background()); err != nil {
		return nil, err
	}
	return app.NewSyncService(client, db, fp), nil
}

func createDirIfNotExist() (string, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	dir := homeDir + "\\.keeper"
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
	}
	return dir, fmt.Sprintf("%s\\keeper.db", dir), err
}
