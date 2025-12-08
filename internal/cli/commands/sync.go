package commands

import (
	"fmt"

	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/spf13/cobra"
)

func BindSyncCommand(root *cobra.Command, userService *app.UserService, syncService *app.SyncService) error {
	var key string
	var pull bool
	var push bool
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync dataManager",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			if syncService != nil {
				return syncService.Sync(ctx, &app.SyncOption{
					PushOnly: push,
					PullOnly: pull,
				})
			}
			fmt.Println("remote server unavailable or not configured")
			return nil
		},
	}
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().BoolVarP(&push, "push", "p", false, "push")
	cmd.Flags().BoolVarP(&pull, "pull", "f", false, "pull")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}
