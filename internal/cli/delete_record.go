package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

type DeleteRecordCommandBuilder struct {
	users    *UserService
	data     *DataService
	sync     *SyncService
	id       string
	key      string
	needSync bool
}

func NewDeleteRecordCommandBuilder(users *UserService, data *DataService, sync *SyncService) *DeleteRecordCommandBuilder {
	return &DeleteRecordCommandBuilder{
		users: users,
		data:  data,
		sync:  sync,
	}
}

func (ctc *DeleteRecordCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "delete-record",
		Short: "delete record",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := ctc.users.Auth(ctx, ctc.key)
			if err != nil {
				return err
			}
			ctx = SetMasterKey(ctx, masterKey)
			if err = ctc.data.DeleteRecord(ctx, ctc.id); err != nil {
				return err
			}
			fmt.Printf("record deleted: %s\n", ctc.id)
			if ctc.needSync && ctc.sync != nil {
				if err = ctc.sync.Sync(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&ctc.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&ctc.key, "key", "k", "", "key")
	cmd.Flags().BoolVarP(&ctc.needSync, "sync", "s", true, "sync")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "id"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return nil, err
	}
	return cmd, nil
}
