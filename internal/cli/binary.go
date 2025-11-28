package cli

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/spf13/cobra"
	"os"
)

type CreateBinaryFileCommandBuilder struct {
	users    *UserService
	data     *DataService
	sync     *SyncService
	key      string
	path     string
	needSync bool
}

func NewCreateBinaryFileCommandBuilder(users *UserService, data *DataService, sync *SyncService) *CreateBinaryFileCommandBuilder {
	return &CreateBinaryFileCommandBuilder{
		users: users,
		data:  data,
		sync:  sync,
	}
}

func (cbf *CreateBinaryFileCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "create-binary-file",
		Short: "Create a binary file for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := cbf.users.Auth(ctx, cbf.key)
			if err != nil {
				return err
			}
			ctx = SetMasterKey(ctx, masterKey)
			id, err := cbf.data.CreateRecord(ctx, &RecordRequest{Type: core.OtherType, Path: cbf.path})
			if err != nil {
				return err
			}
			fmt.Printf("created binary: %s\n", id)
			if cbf.needSync && cbf.sync != nil {
				if err = cbf.sync.Sync(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&cbf.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&cbf.path, "path", "p", "", "path for file")
	cmd.Flags().BoolVarP(&cbf.needSync, "sync", "s", true, "sync")
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
	users    *UserService
	data     *DataService
	sync     *SyncService
	id       string
	key      string
	path     string
	needSync bool
}

func NewUpdateBinaryFileCommandBuilder(users *UserService, data *DataService, sync *SyncService) *UpdateBinaryFileCommandBuilder {
	return &UpdateBinaryFileCommandBuilder{
		users: users,
		data:  data,
		sync:  sync,
	}
}

func (cbf *UpdateBinaryFileCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "update-binary-file",
		Short: "update a binary file for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := cbf.users.Auth(ctx, cbf.key)
			if err != nil {
				return err
			}
			ctx = SetMasterKey(ctx, masterKey)
			id, err := cbf.data.UpdateRecord(ctx, cbf.id, &RecordRequest{Type: core.OtherType, Path: cbf.path})
			if err != nil {
				return err
			}
			fmt.Printf("updated binary: %s\n", id)
			if cbf.needSync && cbf.sync != nil {
				if err = cbf.sync.Sync(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&cbf.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&cbf.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&cbf.path, "path", "p", "", "path for file")
	cmd.Flags().BoolVarP(&cbf.needSync, "sync", "s", true, "sync")
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
	users *UserService
	data  *DataService
	id    string
	key   string
	path  string
}

func NewReadBinaryFileCommandBuilder(users *UserService, data *DataService) *ReadBinaryFileCommandBuilder {
	return &ReadBinaryFileCommandBuilder{
		users: users,
		data:  data,
	}
}

func (cbf *ReadBinaryFileCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "read-binary-file",
		Short: "read a binary file for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := cbf.users.Auth(ctx, cbf.key)
			if err != nil {
				return err
			}
			ctx = SetMasterKey(ctx, masterKey)
			record, err := cbf.data.Get(ctx, cbf.id)
			if err != nil {
				return err
			}
			if record.BigData {
				d, f, err := cbf.data.ExtractFile(ctx, record)
				if err != nil {
					return err
				}
				defer f.Close()
				file, err := os.Create(fmt.Sprintf("%s\\%s", cbf.path, d.Name))
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

	cmd.Flags().StringVarP(&cbf.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&cbf.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&cbf.path, "path", "p", "", "path for file")
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
