package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

type SyncCommandBuilder struct {
	users *UserService
	sync  *SyncService
	key   string
}

func NewSyncCommandBuilder(users *UserService, sync *SyncService) *SyncCommandBuilder {
	return &SyncCommandBuilder{
		users: users,
		sync:  sync,
	}
}

func (scb *SyncCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync data",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := scb.users.Auth(ctx, scb.key)
			if err != nil {
				return err
			}
			ctx = SetMasterKey(ctx, masterKey)
			if scb.sync != nil {
				return scb.sync.Sync(ctx)
			}
			fmt.Println("remote server unavailable or not configured")
			return nil
		},
	}
	cmd.Flags().StringVarP(&scb.key, "key", "k", "", "key")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return nil, err
	}
	return cmd, nil
}
