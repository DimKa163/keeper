package _interface

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/spf13/cobra"
	"os"
)

type CreateBinaryFileCommandBuilder struct {
	userService *app.UserService
	dataManager *app.DataManager
	key         string
	path        string
	needSync    bool
}

func NewCreateBinaryFileCommandBuilder(
	users *app.UserService,
	dataManager *app.DataManager,
) *CreateBinaryFileCommandBuilder {
	return &CreateBinaryFileCommandBuilder{
		userService: users,
		dataManager: dataManager,
	}
}

func (c *CreateBinaryFileCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "create-binary-file",
		Short: "Create a binary file for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.userService.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err := c.dataManager.Create(
				ctx,
				&app.RecordRequest{Type: core.OtherType, Path: c.path},
				c.needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("created binary: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&c.path, "path", "p", "", "path for file")
	cmd.Flags().BoolVarP(&c.needSync, "syncService", "s", true, "syncService")
	err := cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return nil, err
	}
	err = cobra.MarkFlagRequired(cmd.Flags(), "path")
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

type UpdateBinaryFileCommandBuilder struct {
	users    *app.UserService
	data     *app.DataManager
	id       string
	key      string
	path     string
	needSync bool
}

func NewUpdateBinaryFileCommandBuilder(
	users *app.UserService,
	dataManager *app.DataManager,
) *UpdateBinaryFileCommandBuilder {
	return &UpdateBinaryFileCommandBuilder{
		users: users,
		data:  dataManager,
	}
}

func (c *UpdateBinaryFileCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "update-binary-file",
		Short: "update a binary file for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.users.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err := c.data.Update(
				ctx,
				c.id,
				&app.RecordRequest{Type: core.OtherType, Path: c.path},
				c.needSync,
			)
			if err != nil {
				return err
			}
			fmt.Printf("updated binary: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&c.path, "path", "p", "", "path for file")
	cmd.Flags().BoolVarP(&c.needSync, "syncService", "s", true, "syncService")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "id"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "path"); err != nil {
		return nil, err
	}
	return cmd, nil
}

type ReadBinaryFileCommandBuilder struct {
	userService *app.UserService
	dataManager *app.DataManager
	id          string
	key         string
	path        string
}

func NewReadBinaryFileCommandBuilder(users *app.UserService, data *app.DataManager) *ReadBinaryFileCommandBuilder {
	return &ReadBinaryFileCommandBuilder{
		userService: users,
		dataManager: data,
	}
}

func (c *ReadBinaryFileCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "read-binary-file",
		Short: "read a binary file for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.userService.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			record, err := c.dataManager.Get(ctx, c.id)
			if err != nil {
				return err
			}
			if record.BigData {
				d, f, err := c.dataManager.ExtractFile(ctx, record)
				if err != nil {
					return err
				}
				defer f.Close()
				file, err := os.Create(fmt.Sprintf("%s\\%s", c.path, d.Name))
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

	cmd.Flags().StringVarP(&c.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&c.path, "path", "p", "", "path for file")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "id"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "path"); err != nil {
		return nil, err
	}
	return cmd, nil
}
