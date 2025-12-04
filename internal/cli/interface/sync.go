package _interface

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/spf13/cobra"
)

type SyncCommandBuilder struct {
	userService *app.UserService
	syncService *app.SyncService
	key         string
	pull        bool
	push        bool
}

func NewSyncCommandBuilder(userService *app.UserService, syncService *app.SyncService) *SyncCommandBuilder {
	return &SyncCommandBuilder{
		userService: userService,
		syncService: syncService,
	}
}

func (scb *SyncCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "syncService",
		Short: "Sync dataManager",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := scb.userService.Auth(ctx, scb.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			if scb.syncService != nil {
				return scb.syncService.Sync(ctx, &app.SyncOption{
					PushOnly: scb.push,
					PullOnly: scb.pull,
				})
			}
			fmt.Println("remote server unavailable or not configured")
			return nil
		},
	}
	cmd.Flags().StringVarP(&scb.key, "key", "k", "", "key")
	cmd.Flags().BoolVarP(&scb.push, "push", "p", false, "push")
	cmd.Flags().BoolVarP(&scb.pull, "pull", "f", false, "pull")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return nil, err
	}
	return cmd, nil
}
