package _interface

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/spf13/cobra"
)

type CreateTextContentCommandBuilder struct {
	userService *app.UserService
	dataManager *app.DataManager
	key         string
	name        string
	content     string
	needSync    bool
}

func NewCreateTextContentBuilder(userService *app.UserService, dataManager *app.DataManager) *CreateTextContentCommandBuilder {
	return &CreateTextContentCommandBuilder{
		userService: userService,
		dataManager: dataManager,
	}
}

func (c *CreateTextContentCommandBuilder) Build() (*cobra.Command, error) {
	var err error
	cmd := &cobra.Command{
		Use:   "create-text-content",
		Short: "Create text content",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.userService.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err := c.dataManager.Create(
				ctx,
				&app.RecordRequest{Type: core.TextType, Name: c.name, Content: c.content},
				c.needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("created text: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&c.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&c.content, "content", "c", "", "content")
	cmd.Flags().BoolVarP(&c.needSync, "syncService", "s", true, "syncService")
	err = cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return nil, err
	}
	return cmd, nil
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

func (c *UpdateTextContentCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "update-text-content",
		Short: "Update text content",
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
				&app.RecordRequest{Type: core.TextType, Name: c.name, Content: c.content},
				c.needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("updated text: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&c.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&c.content, "content", "c", "", "content")
	cmd.Flags().BoolVarP(&c.needSync, "syncService", "s", true, "syncService")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "id"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return nil, err
	}
	return cmd, nil
}
