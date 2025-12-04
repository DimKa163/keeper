package _interface

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/spf13/cobra"
)

type DeleteRecordCommandBuilder struct {
	userService *app.UserService
	dataManager *app.DataManager
	id          string
	key         string
	needSync    bool
}

func NewDeleteRecordCommandBuilder(
	userService *app.UserService,
	dataManager *app.DataManager,
) *DeleteRecordCommandBuilder {
	return &DeleteRecordCommandBuilder{
		userService: userService,
		dataManager: dataManager,
	}
}

func (c *DeleteRecordCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "delete-record",
		Short: "delete record",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.userService.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			if err = c.dataManager.Delete(ctx, c.id, c.needSync); err != nil {
				return err
			}
			fmt.Printf("record deleted: %s\n", c.id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().BoolVarP(&c.needSync, "syncService", "s", true, "syncService")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "id"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return nil, err
	}
	return cmd, nil
}
