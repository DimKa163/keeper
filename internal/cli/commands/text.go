package commands

import (
	"context"
	"fmt"

	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/spf13/cobra"
)

type TextManager interface {
	CreateText(ctx context.Context, request *app.TextRequest, sync bool) (string, error)
	UpdateText(ctx context.Context, id string, request *app.TextRequest, sync bool) (string, error)
}

func BindCreateTextCommand(root *cobra.Command, userService *app.UserService, dataManager TextManager) error {
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
			id, err := dataManager.CreateText(
				ctx,
				&app.TextRequest{Name: name, Content: content},
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

func BindUpdateTextCommand(root *cobra.Command, userService *app.UserService, dataManager TextManager) error {
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
			id, err = dataManager.UpdateText(
				ctx,
				id,
				&app.TextRequest{Name: name, Content: content},
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
