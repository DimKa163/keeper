package cli

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/spf13/cobra"
)

type CreateLoginPassCommandBuilder struct {
	users    *UserService
	data     *DataService
	sync     *SyncService
	key      string
	name     string
	login    string
	pass     string
	url      string
	needSync bool
}

func NewCreateLoginPassBuilder(users *UserService, data *DataService, sync *SyncService) *CreateLoginPassCommandBuilder {
	return &CreateLoginPassCommandBuilder{
		users: users,
		data:  data,
		sync:  sync,
	}
}

func (clp *CreateLoginPassCommandBuilder) Build() (*cobra.Command, error) {
	var err error
	cmd := &cobra.Command{
		Use:   "create-login-pass",
		Short: "Create login pass",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := clp.users.Auth(ctx, clp.key)
			if err != nil {
				return err
			}
			ctx = SetMasterKey(ctx, masterKey)
			id, err := clp.data.CreateRecord(ctx, &RecordRequest{Type: core.LoginPassType, Name: clp.name, Login: clp.login, Pass: clp.pass, Url: clp.url})
			if err != nil {
				return err
			}
			fmt.Printf("Created login pass: %s\n", id)
			if clp.needSync && clp.sync != nil {
				if err = clp.sync.Sync(ctx); err != nil {
					fmt.Printf("Failed to sync login pass: %s\n", id)
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&clp.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&clp.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&clp.login, "login", "l", "", "login name")
	cmd.Flags().StringVarP(&clp.pass, "pass", "p", "", "pass")
	cmd.Flags().StringVarP(&clp.url, "url", "u", "", "url")
	cmd.Flags().BoolVarP(&clp.needSync, "sync", "s", true, "sync")
	err = cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return nil, err
	}
	err = cobra.MarkFlagRequired(cmd.Flags(), "pass")
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

type UpdateLoginPassCommandBuilder struct {
	users    *UserService
	data     *DataService
	sync     *SyncService
	id       string
	key      string
	name     string
	login    string
	pass     string
	url      string
	needSync bool
}

func NewUpdateLoginPassCommandBuilder(users *UserService, data *DataService, sync *SyncService) *UpdateLoginPassCommandBuilder {
	return &UpdateLoginPassCommandBuilder{
		users: users,
		data:  data,
		sync:  sync,
	}
}

func (clp *UpdateLoginPassCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "update-login-pass",
		Short: "Update login pass",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := clp.users.Auth(ctx, clp.key)
			if err != nil {
				return err
			}
			ctx = SetMasterKey(ctx, masterKey)
			id, err := clp.data.UpdateRecord(ctx, clp.id, &RecordRequest{Type: core.LoginPassType, Name: clp.name, Login: clp.login, Pass: clp.pass, Url: clp.url})
			if err != nil {
				return err
			}
			fmt.Printf("updated login pass: %s\n", id)
			if clp.needSync && clp.sync != nil {
				if err = clp.sync.Sync(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&clp.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&clp.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&clp.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&clp.login, "login", "l", "", "login name")
	cmd.Flags().StringVarP(&clp.pass, "pass", "p", "", "pass")
	cmd.Flags().StringVarP(&clp.url, "url", "u", "", "url")
	cmd.Flags().BoolVarP(&clp.needSync, "sync", "s", true, "sync")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "id"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "pass"); err != nil {
		return nil, err
	}
	return cmd, nil
}
