package cli

import (
	"github.com/spf13/cobra"
)

type RegisterBuilder struct {
	users *UserService
	key   string
}

func NewRegisterBuilder(users *UserService) *RegisterBuilder {
	return &RegisterBuilder{users: users}
}

func (rb *RegisterBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "keeper reg-master-key",
		Short: "create a master key",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return rb.users.Register(ctx, rb.key)
		},
	}
	cmd.Flags().StringVarP(&rb.key, "key", "k", "", "key")
	err := cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return nil, err
	}
	return cmd, nil
}
