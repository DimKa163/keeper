package commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/spf13/cobra"
)

func BindConflictSolveCommand(root *cobra.Command, dataManager *app.DataManager, userService *app.UserService, syncService *app.SyncService) error {
	var key string
	cmd := &cobra.Command{
		Use: "resolve-conflict",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			conflicts, err := dataManager.GetAllConflicts(ctx)
			if err != nil {
				return err
			}
			for _, conflict := range conflicts {
				if err = printMenu(ctx, dataManager, conflict); err != nil {
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

				if err = dataManager.SolveConflict(ctx, app.Version(i), conflict); err != nil {
					return err
				}

				if err = dataManager.DeleteConflict(ctx, conflict); err != nil {
					return err
				}
			}
			if err = syncService.Sync(ctx, &app.SyncOption{
				Force: true,
			}); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "key"); err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
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
