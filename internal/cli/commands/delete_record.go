package commands

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/spf13/cobra"
)

func BindDeleteCommand(root *cobra.Command, userService *app.UserService, dataManager *app.DataManager) error {
	var id string
	var key string
	var needSync bool
	cmd := &cobra.Command{
		Use:   "delete-record",
		Short: "delete record",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterkey, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterkey)
			if err = dataManager.Delete(ctx, id, needSync); err != nil {
				return err
			}
			fmt.Printf("record deleted: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().BoolVarP(&needSync, "syncService", "s", true, "syncService")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "id"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}
