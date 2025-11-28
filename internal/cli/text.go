package cli

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/spf13/cobra"
)

type CreateTextContentCommandBuilder struct {
	users    *UserService
	data     *DataService
	sync     *SyncService
	key      string
	name     string
	content  string
	needSync bool
}

func NewCreateTextContentBuilder(users *UserService, data *DataService, sync *SyncService) *CreateTextContentCommandBuilder {
	return &CreateTextContentCommandBuilder{
		users: users,
		data:  data,
		sync:  sync,
	}
}

func (ctc *CreateTextContentCommandBuilder) Build() (*cobra.Command, error) {
	var err error
	cmd := &cobra.Command{
		Use:   "create-text-content",
		Short: "Create text content",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := ctc.users.Auth(ctx, ctc.key)
			if err != nil {
				return err
			}
			ctx = SetMasterKey(ctx, masterKey)
			id, err := ctc.data.CreateRecord(ctx, &RecordRequest{Type: core.TextType, Name: ctc.name, Content: ctc.content})
			if err != nil {
				return err
			}
			fmt.Printf("created text: %s\n", id)
			if ctc.needSync && ctc.sync != nil {
				if err = ctc.sync.Sync(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&ctc.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&ctc.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&ctc.content, "content", "c", "", "content")
	cmd.Flags().BoolVarP(&ctc.needSync, "sync", "s", true, "sync")
	err = cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

type UpdateTextContentCommandBuilder struct {
	users    *UserService
	data     *DataService
	sync     *SyncService
	id       string
	key      string
	name     string
	content  string
	needSync bool
}

func NewUpdateTextContentCommandBuilder(users *UserService, data *DataService, sync *SyncService) *UpdateTextContentCommandBuilder {
	return &UpdateTextContentCommandBuilder{
		users: users,
		data:  data,
		sync:  sync,
	}
}

func (ctc *UpdateTextContentCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "update-text-content",
		Short: "Update text content",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := ctc.users.Auth(ctx, ctc.key)
			if err != nil {
				return err
			}
			ctx = SetMasterKey(ctx, masterKey)
			id, err := ctc.data.UpdateRecord(ctx, ctc.id, &RecordRequest{Type: core.TextType, Name: ctc.name, Content: ctc.content})
			if err != nil {
				return err
			}
			fmt.Printf("updated text: %s\n", id)
			if ctc.needSync && ctc.sync != nil {
				if err = ctc.sync.Sync(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&ctc.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&ctc.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&ctc.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&ctc.content, "content", "c", "", "content")
	cmd.Flags().BoolVarP(&ctc.needSync, "sync", "s", true, "sync")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "id"); err != nil {
		return nil, err
	}
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return nil, err
	}
	return cmd, nil
}
