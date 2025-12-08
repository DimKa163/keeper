package commands

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/spf13/cobra"
)

func BindCreateTextCommand(root *cobra.Command, userService *app.UserService, dataManager *app.DataManager) error {
	var key string
	var name string
	var content string
	var needSync bool
	cmd := &cobra.Command{
		Use:   "create-text-content",
		Short: "Create text content",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err := dataManager.Create(
				ctx,
				&app.RecordRequest{Type: core.TextType, Name: name, Content: content},
				needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("created text: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&content, "content", "c", "", "content")
	cmd.Flags().BoolVarP(&needSync, "syncService", "s", true, "syncService")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}

type UpdateTextContentCommandBuilder struct {
	userService *app.UserService
	dataManager *app.DataManager
	id          string
	key         string
	name        string
	content     string
	needSync    bool
}

func NewUpdateTextContentCommandBuilder(userService *app.UserService, dataManager *app.DataManager) *UpdateTextContentCommandBuilder {
	return &UpdateTextContentCommandBuilder{
		userService: userService,
		dataManager: dataManager,
	}
}

func BindUpdateTextCommand(root *cobra.Command, userService *app.UserService, dataManager *app.DataManager) error {
	var key string
	var id string
	var name string
	var content string
	var needSync bool
	cmd := &cobra.Command{
		Use:   "update-text-content",
		Short: "Update text content",
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
				&app.RecordRequest{Type: core.TextType, Name: name, Content: content},
				needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("updated text: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&content, "content", "c", "", "content")
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
