package commands

import (
	"fmt"

	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/spf13/cobra"
)

func BindListCommand(root *cobra.Command, userService *app.UserService, dataManager *app.DataManager) error {
	var key string
	var limit int32
	var offset int32
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all dataManager",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			key, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, key)
			records, err := dataManager.GetAll(ctx, limit, offset)
			if err != nil {
				return err
			}
			var js string
			for _, record := range records {
				var data []byte
				data, err = dataManager.Decode(ctx, record)
				if err != nil {
					return err
				}
				js, err = toViewModel(record, data)
				if err != nil {
					return err
				}
				fmt.Println(js)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().Int32VarP(&limit, "limit", "l", 5, "limit")
	cmd.Flags().Int32VarP(&offset, "offset", "o", 0, "offset")
	err := cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}
