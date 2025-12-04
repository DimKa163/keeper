package _interface

import (
	"database/sql"
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/spf13/cobra"
)

type RegisterRemoteServerCommandBuilder struct {
	userService *app.UserService
	db          *sql.DB
	key         string
	addr        string
	login       string
	pass        string
	active      bool
}

func NewRegisterRemoteServerCommandBuilder(userService *app.UserService, db *sql.DB) *RegisterRemoteServerCommandBuilder {
	return &RegisterRemoteServerCommandBuilder{
		userService: userService,
		db:          db,
	}
}

func (c *RegisterRemoteServerCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "register-remote-server",
		Short: "register remote server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.userService.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			if err = persistence.InsertServer(ctx, c.db, c.addr, c.login, c.pass, c.active); err != nil {
				return err
			}
			client, err := app.NewRemoteClient(c.addr, c.login, c.pass)
			if err != nil {
				return err
			}
			if err = client.IsHealthy(ctx); err != nil {
				return err
			}
			if err = client.TryToAuthenticate(ctx); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&c.addr, "addr", "a", "", "server listen address")
	cmd.Flags().StringVarP(&c.login, "login", "l", "", "remote server login")
	cmd.Flags().StringVarP(&c.pass, "pass", "p", "", "remote server password")
	cmd.Flags().BoolVarP(&c.active, "active", "i", true, "active remote server")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "addr"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "login"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "pass"); err != nil {
		return nil, err
	}
	return cmd, nil
}
