package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/DimKa163/keeper/internal/cli"
	"github.com/DimKa163/keeper/internal/cli/crypto"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	_ "modernc.org/sqlite"
)

type CMD struct {
	*cli.ServiceContainer
	commands map[string]cli.CommandHandler
}

func New() (*CMD, error) {
	path, err := createDirIfNotExist()
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

	recordRep := persistence.NewRecordRepository(db)
	userRep := persistence.NewUserRepository(db)

	return &CMD{
		ServiceContainer: &cli.ServiceContainer{
			DB:               db,
			RecordRepository: recordRep,
			UserRepository:   userRep,
			UserService:      cli.NewUserService(userRep),
			DataService:      cli.NewDataService(db),
			Encoder:          crypto.NewAesEncoder(),
			Decoder:          crypto.NewAesDecoder(),
			Console:          cli.NewConsole(),
		},
		commands: make(map[string]cli.CommandHandler),
	}, nil
}

func (cmd *CMD) AddCommand(name string, handler cli.CommandHandler) {
	cmd.commands[name] = handler
}

func (cmd *CMD) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	console := cmd.Console
	masterKey, err := cmd.UserService.Init(ctx, console)
	if err != nil {
		return err
	}
	cliCtx := cli.NewCLI(ctx, cmd.ServiceContainer, masterKey)
	commandChan := make(chan cli.Command)
	defer close(commandChan)
	go func() {
		for {
			fmt.Print("keeper>")
			rawCommand, err := console.Read()
			if err != nil {
				fmt.Println("Error: ", err)
				continue
			}
			commandChan <- cli.Parse(rawCommand)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case command := <-commandChan:
			if err = command.Determinate(); err != nil {
				fmt.Println("Error: ", err)
				continue
			}
			if command.Name() == "q" {
				break
			}
			handler, ok := cmd.commands[command.Name()]
			if !ok {
				fmt.Println("Error: command not found", command)
				continue
			}
			if err := handler(cliCtx); err != nil {
				fmt.Println("Error: ", err)
				continue
			}
		}
	}
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
	}
	return fmt.Sprintf("%s\\keeper.db", dir), err
}
