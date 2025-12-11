// Package cmd start up module
package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/commands"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/crypto"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/DimKa163/keeper/internal/datatool"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

type ServiceContainer struct {
	DB          *sql.DB
	UserService *app.UserService
	SyncService app.Syncer
	DataService *app.DataManager
	Decoder     core.Decoder
	Encoder     core.Encoder
}

type CMD struct {
	*ServiceContainer
	root    *cobra.Command
	version string
	commit  string
	date    string
}

func New(version, commit, date string) (*CMD, error) {

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
	fileProvider := datatool.NewFileProvider(fmt.Sprintf("%s\\", dir))
	syncService, err := createSyncService(db, fileProvider)
	if err != nil && errors.Is(err, app.ErrServerUnavailable) {
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
		version: version,
		commit:  commit,
		date:    date,
	}

	rootCmd := &cobra.Command{
		Use:  "keeper",
		RunE: cmd.Run,
	}
	cmd.root = rootCmd
	return cmd, nil
}

func (cmd *CMD) RegisterCommands() error {
	if err := commands.BindListCommand(cmd.root, cmd.UserService, cmd.DataService); err != nil {
		return err
	}
	if err := commands.BindRegister(cmd.root, cmd.UserService); err != nil {
		return err
	}
	if err := commands.BindCreateLoginPassCommand(cmd.root, cmd.UserService, cmd.DataService); err != nil {
		return err
	}
	if err := commands.BindUpdateLoginPassCommand(cmd.root, cmd.UserService, cmd.DataService); err != nil {
		return err
	}
	if err := commands.BindCreateTextCommand(cmd.root, cmd.UserService, cmd.DataService); err != nil {
		return err
	}
	if err := commands.BindUpdateTextCommand(cmd.root, cmd.UserService, cmd.DataService); err != nil {
		return err
	}
	if err := commands.BindCreateBinary(cmd.root, cmd.UserService, cmd.DataService); err != nil {
		return err
	}
	if err := commands.BindUpdateBinary(cmd.root, cmd.UserService, cmd.DataService); err != nil {
		return err
	}
	if err := commands.BindCreateBankCard(cmd.root, cmd.UserService, cmd.DataService); err != nil {
		return err
	}
	if err := commands.BindUpdateBankCard(cmd.root, cmd.UserService, cmd.DataService); err != nil {
		return err
	}
	if err := commands.BindDeleteCommand(cmd.root, cmd.UserService, cmd.DataService); err != nil {
		return err
	}
	if err := commands.BindRegisterRemoteServer(cmd.root, cmd.UserService, cmd.DB); err != nil {
		return err
	}
	if err := commands.BindSyncCommand(cmd.root, cmd.UserService, cmd.SyncService); err != nil {
		return err
	}
	if err := commands.BindExportBinaryCommand(cmd.root, cmd.UserService, cmd.DataService); err != nil {
		return err
	}
	if err := commands.BindConflictSolveCommand(cmd.root, cmd.DataService, cmd.UserService, cmd.SyncService); err != nil {
		return err
	}
	if err := commands.BindGetVersionCommand(cmd.root, cmd.version, cmd.commit, cmd.date); err != nil {
		return err
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
	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Second)
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

func createSyncService(db *sql.DB, fp *datatool.FileProvider) (app.Syncer, error) {
	serv, err := persistence.GetServer(context.Background(), db, true)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return app.NewEmptySyncer(), nil
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
