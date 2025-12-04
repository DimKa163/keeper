package _interface

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/spf13/cobra"
	"strconv"
)

type ResolveCommandBuilder struct {
	dataManager *app.DataManager
	users       *app.UserService
	syncService *app.SyncService
	key         string
}

func NewResolveCommandBuilder(dataManager *app.DataManager, user *app.UserService, syncService *app.SyncService) *ResolveCommandBuilder {
	return &ResolveCommandBuilder{
		dataManager: dataManager,
		users:       user,
		syncService: syncService,
	}
}

func (c *ResolveCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use: "resolve-conflict",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.users.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			conflicts, err := c.dataManager.GetAllConflicts(ctx)
			if err != nil {
				return err
			}
			for _, conflict := range conflicts {
				if err = printMenu(ctx, c.dataManager, conflict); err != nil {
					return err
				}

				var input string
				if _, err = fmt.Scanln(&input); err != nil {
					return err
				}

				i, err := strconv.Atoi(input)
				if err != nil {
					var numErr *strconv.NumError
					if errors.As(err, &numErr) {
						fmt.Printf("%s not a number\n", input)
						continue
					}
					return err
				}

				if err = c.dataManager.SolveConflict(ctx, app.Version(i), conflict); err != nil {
					return err
				}

				if err = c.dataManager.DeleteConflict(ctx, conflict); err != nil {
					return err
				}
			}
			if err = c.syncService.Sync(ctx, &app.SyncOption{
				Force: true,
			}); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return nil, err
	}
	return cmd, nil
}

func printMenu(ctx context.Context, manager *app.DataManager, conflict *core.Conflict) error {
	local := conflict.Local
	remote := conflict.Remote
	fmt.Println("1. local:")
	if err := printSecret(ctx, manager, local.Record); err != nil {
		return err
	}
	fmt.Println("2. remote:")
	if err := printSecret(ctx, manager, remote.Record); err != nil {
		return err
	}
	fmt.Println("Local:")
	fmt.Println("Select:")
	fmt.Println("1 - local")
	fmt.Println("2 - remote")
	return nil
}
func printSecret(ctx context.Context, manager *app.DataManager, record *core.Record) error {
	data, err := manager.Decode(ctx, record)
	if err != nil {
		return err
	}
	js, err := toViewModel(record, data)
	if err != nil {
		return err
	}
	fmt.Println(js)
	return nil
}
func toViewModel(record *core.Record, data []byte) (string, error) {
	switch record.Type {
	case core.LoginPassType:
		var loginPass core.LoginPass
		if err := json.Unmarshal(data, &loginPass); err != nil {
			return "", err
		}
		return mapLoginPass(record.ID, &loginPass)
	case core.TextType:
		var text core.Text
		if err := json.Unmarshal(data, &text); err != nil {
			return "", err
		}
		return mapText(record.ID, &text)
	case core.BankCardType:
		var bankCard core.BankCard
		if err := json.Unmarshal(data, &bankCard); err != nil {
			return "", err
		}
		return mapBakCard(record.ID, &bankCard)
	case core.OtherType:
		var binary core.Binary
		if err := json.Unmarshal(data, &binary); err != nil {
			return "", err
		}
		return mapOther(record.ID, &binary)
	}
	return "", errors.New("invalid record")
}
