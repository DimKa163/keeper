package commands

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

func BindCreateLoginPassCommand(root *cobra.Command, userService *app.UserService, dataManager *app.DataManager) error {
	var key string
	var name string
	var login string
	var pass string
	var url string
	var needSync bool
	var err error
	cmd := &cobra.Command{
		Use:   "create-login-pass",
		Short: "Create login pass",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err := dataManager.Create(
				ctx,
				&app.RecordRequest{Type: core.LoginPassType, Name: name, Login: login, Pass: pass, Url: url},
				needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("Created login pass: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&login, "login", "l", "", "login name")
	cmd.Flags().StringVarP(&pass, "pass", "p", "", "pass")
	cmd.Flags().StringVarP(&url, "url", "u", "", "url")
	cmd.Flags().BoolVarP(&needSync, "syncService", "s", true, "syncService")
	err = cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return err
	}
	err = cobra.MarkFlagRequired(cmd.Flags(), "pass")
	if err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}

func BindUpdateLoginPassCommand(root *cobra.Command, userService *app.UserService, dataManager *app.DataManager) error {
	var key string
	var id string
	var name string
	var login string
	var pass string
	var url string
	var needSync bool
	cmd := &cobra.Command{
		Use:   "update-login-pass",
		Short: "Update login pass",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err := dataManager.Update(
				ctx,
				id,
				&app.RecordRequest{Type: core.LoginPassType, Name: name, Login: login, Pass: pass, Url: url},
				needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("updated login pass: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&login, "login", "l", "", "login name")
	cmd.Flags().StringVarP(&pass, "pass", "p", "", "pass")
	cmd.Flags().StringVarP(&url, "url", "u", "", "url")
	cmd.Flags().BoolVarP(&needSync, "syncService", "s", true, "syncService")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "id"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "pass"); err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}
