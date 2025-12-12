package commands

import (
	"database/sql"

	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/persistence"
	"github.com/spf13/cobra"
)

func BindRegisterRemoteServer(root *cobra.Command, userService *app.UserService, db *sql.DB) error {
	var key string
	var addr string
	var login string
	var pass string
	var active bool
	cmd := &cobra.Command{
		Use:   "register-remote-server",
		Short: "register remote server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			if err = persistence.InsertServer(ctx, db, addr, login, pass, active); err != nil {
				return err
			}
			client, err := app.NewRemoteClient(addr, login, pass)
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
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&addr, "addr", "a", "", "server listen address")
	cmd.Flags().StringVarP(&login, "login", "l", "", "remote server login")
	cmd.Flags().StringVarP(&pass, "pass", "p", "", "remote server password")
	cmd.Flags().BoolVarP(&active, "active", "i", true, "active remote server")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "addr"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "login"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "pass"); err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}
