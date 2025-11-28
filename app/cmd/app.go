package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DimKa163/keeper/internal/cli"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/crypto"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/DimKa163/keeper/internal/shared"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
	"os"
	"reflect"
	"time"
)

type CMD struct {
	*cli.ServiceContainer
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
	if err = persistence.Migrate(db, "./internal/cli/migrations"); err != nil {
		return nil, err
	}
	fileProvider := shared.NewFileProvider(fmt.Sprintf("%s\\", dir))
	syncService, err := createSyncService(db, fileProvider)
	if err != nil && errors.Is(cli.ErrServerUnavailable, err) {
		fmt.Println("remote server unavailable")
		err = nil
	}
	if err != nil {
		return nil, err
	}
	encoder := crypto.NewGzipEncoder(crypto.NewAesEncoder())
	decoder := crypto.NewGzipDecoder(crypto.NewAesDecoder())
	app := &CMD{
		ServiceContainer: &cli.ServiceContainer{
			DB:          db,
			UserService: cli.NewUserService(db),
			DataService: cli.NewDataService(db, encoder, decoder, fileProvider),
			SyncService: syncService,
			Encoder:     encoder,
			Decoder:     decoder,
		},
	}

	rootCmd := &cobra.Command{
		Use:  "keeper",
		RunE: app.Run,
	}
	app.root = rootCmd
	return app, nil
}

func (cmd *CMD) RegisterCommands() error {
	if err := cmd.addCommand(cli.NewRegisterBuilder(cmd.UserService)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewDataListBuilder(cmd.UserService, cmd.DataService, cmd.Decoder)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewCreateLoginPassBuilder(cmd.UserService, cmd.DataService, cmd.SyncService)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewUpdateLoginPassCommandBuilder(cmd.UserService, cmd.DataService, cmd.SyncService)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewCreateTextContentBuilder(cmd.UserService, cmd.DataService, cmd.SyncService)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewUpdateTextContentCommandBuilder(cmd.UserService, cmd.DataService, cmd.SyncService)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewCreateBinaryFileCommandBuilder(cmd.UserService, cmd.DataService, cmd.SyncService)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewUpdateBinaryFileCommandBuilder(cmd.UserService, cmd.DataService, cmd.SyncService)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewCreateBankCardCommandBuilder(cmd.UserService, cmd.DataService, cmd.SyncService)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewUpdateBankCardCommandBuilder(cmd.UserService, cmd.DataService, cmd.SyncService)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewDeleteRecordCommandBuilder(cmd.UserService, cmd.DataService, cmd.SyncService)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewRegisterRemoteServerCommandBuilder(cmd.UserService, cmd.DB)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewSyncCommandBuilder(cmd.UserService, cmd.SyncService)); err != nil {
		return err
	}
	if err := cmd.addCommand(cli.NewReadBinaryFileCommandBuilder(cmd.UserService, cmd.DataService)); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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
		ctx = cli.SetMasterKey(cli.SetHostName(ctx, us.Username), us.Password)
	}
	name := reflect.TypeOf(core.Record{}).Name()
	version, err := persistence.GetState(ctx, cmd.DB, name)
	if err != nil {
		return err
	}
	ctx = cli.SetVersion(ctx, version.Value)
	return cmd.root.ExecuteContext(ctx)
}

func createSyncService(db *sql.DB, fp *shared.FileProvider) (*cli.SyncService, error) {
	serv, err := persistence.GetServer(context.Background(), db, true)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, nil
	}
	client, err := cli.NewRemoteClient(serv.Address, serv.Login, serv.Password)
	if err != nil {
		return nil, err
	}
	if err = client.IsHealthy(context.Background()); err != nil {
		return nil, err
	}
	return cli.NewSyncService(client, db, fp), nil
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
