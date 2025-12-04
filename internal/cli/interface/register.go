package _interface

import (
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/spf13/cobra"
)

type RegisterBuilder struct {
	userService *app.UserService
	key         string
}

func NewRegisterBuilder(userService *app.UserService) *RegisterBuilder {
	return &RegisterBuilder{userService: userService}
}

func (c *RegisterBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "keeper reg-master-key",
		Short: "create a master key",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return c.userService.Register(ctx, c.key)
		},
	}
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	err := cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return nil, err
	}
	return cmd, nil
}
