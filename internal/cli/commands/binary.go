// Package commands interface for cli-application
package commands

import (
	"context"
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/core"
	"io"
	"os"

	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/spf13/cobra"
)

type BinaryManager interface {
	RecordReader
	CreateBinary(ctx context.Context, req *app.BinaryRequest, sync bool) (string, error)
	UpdateBinary(ctx context.Context, id string, req *app.BinaryRequest, sync bool) (string, error)

	ExtractFile(ctx context.Context, record *core.Record) (*core.Binary, io.ReadCloser, error)
}

func BindCreateBinary(root *cobra.Command, userService *app.UserService, dataManager BinaryManager) error {
	var key string
	var path string
	var needSync bool
	cmd := &cobra.Command{
		Use:   "create-binary-file",
		Short: "Create a binary file for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			key, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, key)
			id, err := dataManager.CreateBinary(
				ctx,
				&app.BinaryRequest{Path: path},
				needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("created binary: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&path, "path", "p", "", "path for file")
	cmd.Flags().BoolVarP(&needSync, "syncService", "s", true, "syncService")
	err := cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return err
	}
	err = cobra.MarkFlagRequired(cmd.Flags(), "path")
	if err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}
func BindUpdateBinary(root *cobra.Command, userService *app.UserService, dataManager BinaryManager) error {
	var key string
	var id string
	var path string
	var needSync bool
	cmd := &cobra.Command{
		Use:   "update-binary-file",
		Short: "update a binary file for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err = dataManager.UpdateBinary(
				ctx,
				id,
				&app.BinaryRequest{Path: path},
				needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("updated binary: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&path, "path", "p", "", "path for file")
	cmd.Flags().BoolVarP(&needSync, "syncService", "s", true, "syncService")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "id"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "path"); err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}

func BindExportBinaryCommand(root *cobra.Command, userService *app.UserService, dataManager BinaryManager) error {
	var id string
	var key string
	var path string
	cmd := &cobra.Command{
		Use:   "read-binary-file",
		Short: "read a binary file for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			record, err := dataManager.Get(ctx, id)
			if err != nil {
				return err
			}
			if record.BigData {
				d, f, err := dataManager.ExtractFile(ctx, record)
				if err != nil {
					return err
				}
				defer f.Close()
				file, err := os.Create(fmt.Sprintf("%s\\%s", path, d.Name))
				if err != nil {
					return err
				}
				defer file.Close()
				content := make([]byte, d.SizeBytes)
				_, err = f.Read(content)
				if err != nil {
					return err
				}
				_, err = file.Write(content)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&path, "path", "p", "", "path for file")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "id"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "path"); err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}
