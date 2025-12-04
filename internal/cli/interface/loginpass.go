package _interface

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/spf13/cobra"
)

type CreateLoginPassCommandBuilder struct {
	userService *app.UserService
	dataManager *app.DataManager
	key         string
	name        string
	login       string
	pass        string
	url         string
	needSync    bool
}

func NewCreateLoginPassBuilder(userService *app.UserService, dataManager *app.DataManager) *CreateLoginPassCommandBuilder {
	return &CreateLoginPassCommandBuilder{
		userService: userService,
		dataManager: dataManager,
	}
}

func (c *CreateLoginPassCommandBuilder) Build() (*cobra.Command, error) {
	var err error
	cmd := &cobra.Command{
		Use:   "create-login-pass",
		Short: "Create login pass",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.userService.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err := c.dataManager.Create(
				ctx,
				&app.RecordRequest{Type: core.LoginPassType, Name: c.name, Login: c.login, Pass: c.pass, Url: c.url},
				c.needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("Created login pass: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&c.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&c.login, "login", "l", "", "login name")
	cmd.Flags().StringVarP(&c.pass, "pass", "p", "", "pass")
	cmd.Flags().StringVarP(&c.url, "url", "u", "", "url")
	cmd.Flags().BoolVarP(&c.needSync, "syncService", "s", true, "syncService")
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
	userService *app.UserService
	dataManager *app.DataManager
	id          string
	key         string
	name        string
	login       string
	pass        string
	url         string
	needSync    bool
}

func NewUpdateLoginPassCommandBuilder(userService *app.UserService, dataManager *app.DataManager) *UpdateLoginPassCommandBuilder {
	return &UpdateLoginPassCommandBuilder{
		userService: userService,
		dataManager: dataManager,
	}
}

func (c *UpdateLoginPassCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "update-login-pass",
		Short: "Update login pass",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.userService.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err := c.dataManager.Update(
				ctx,
				c.id,
				&app.RecordRequest{Type: core.LoginPassType, Name: c.name, Login: c.login, Pass: c.pass, Url: c.url},
				c.needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("updated login pass: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&c.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&c.login, "login", "l", "", "login name")
	cmd.Flags().StringVarP(&c.pass, "pass", "p", "", "pass")
	cmd.Flags().StringVarP(&c.url, "url", "u", "", "url")
	cmd.Flags().BoolVarP(&c.needSync, "syncService", "s", true, "syncService")
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
