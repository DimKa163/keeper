package commands

import (
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/spf13/cobra"
)

func BindRegister(root *cobra.Command, userService *app.UserService) error {
	var key string
	cmd := &cobra.Command{
		Use:   "keeper reg-master-key",
		Short: "create a master key",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return userService.Register(ctx, key)
		},
	}
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	err := cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}
