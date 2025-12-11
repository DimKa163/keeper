package commands

import (
	"context"
	"fmt"

	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/spf13/cobra"
)

type BankCardManager interface {
	CreateBankCard(ctx context.Context, req *app.BankCardRequest, sync bool) (string, error)
	UpdateBankCard(ctx context.Context, id string, req *app.BankCardRequest, sync bool) (string, error)
}

func BindCreateBankCard(root *cobra.Command, userService *app.UserService, dataManager BankCardManager) error {
	var key string
	var name string
	var cardNumber string
	var expiry string
	var cvv string
	var holderName string
	var bankName string
	var cardType string
	var currency string
	var isPrimary bool
	var needSync bool
	cmd := &cobra.Command{
		Use:   "create-bank-card",
		Short: "Create a bank card",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err := dataManager.CreateBankCard(ctx, &app.BankCardRequest{
				Name:       name,
				CardNumber: cardNumber,
				Expiry:     expiry,
				CVV:        cvv,
				HolderName: holderName,
				BankName:   bankName,
				CardType:   cardType,
				Currency:   currency,
				IsPrimary:  isPrimary,
			}, needSync)
			if err != nil {
				return err
			}
			fmt.Printf("created bank card dataManager: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&cardNumber, "number", "u", "", "card number")
	cmd.Flags().StringVarP(&expiry, "expire", "e", "", "expire date")
	cmd.Flags().StringVarP(&cvv, "cvv", "v", "", "CVV")
	cmd.Flags().StringVarP(&holderName, "holder-name", "o", "", "holder name")
	cmd.Flags().StringVarP(&bankName, "bank-name", "b", "", "bank name")
	cmd.Flags().StringVarP(&cardType, "card-type", "t", "", "card type")
	cmd.Flags().StringVarP(&currency, "currency", "c", "", "currency")
	cmd.Flags().BoolVarP(&isPrimary, "primary", "p", false, "primary")
	cmd.Flags().BoolVarP(&needSync, "syncService", "s", true, "syncService")
	err := cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return err
	}
	err = cobra.MarkFlagRequired(cmd.Flags(), "number")
	if err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}

func BindUpdateBankCard(root *cobra.Command, userService *app.UserService, dataManager BankCardManager) error {
	var key string
	var id string
	var name string
	var cardNumber string
	var expiry string
	var cvv string
	var holderName string
	var bankName string
	var cardType string
	var currency string
	var isPrimary bool
	var needSync bool
	cmd := &cobra.Command{
		Use:   "update-bank-card",
		Short: "update a bank card dataManager",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := userService.Auth(ctx, key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err = dataManager.UpdateBankCard(ctx, id, &app.BankCardRequest{
				Name:       name,
				CardNumber: cardNumber,
				Expiry:     expiry,
				CVV:        cvv,
				HolderName: holderName,
				BankName:   bankName,
				CardType:   cardType,
				Currency:   currency,
				IsPrimary:  isPrimary,
			}, needSync)
			if err != nil {
				return err
			}
			fmt.Printf("updated bank card dataManager: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&cardNumber, "number", "u", "", "card number")
	cmd.Flags().StringVarP(&expiry, "expire", "e", "", "expire date")
	cmd.Flags().StringVarP(&cvv, "cvv", "v", "", "CVV")
	cmd.Flags().StringVarP(&holderName, "holder-name", "o", "", "holder name")
	cmd.Flags().StringVarP(&bankName, "bank-name", "b", "", "bank name")
	cmd.Flags().StringVarP(&cardType, "card-type", "t", "", "card type")
	cmd.Flags().StringVarP(&currency, "currency", "c", "", "currency")
	cmd.Flags().BoolVarP(&isPrimary, "primary", "p", false, "primary")
	cmd.Flags().BoolVarP(&needSync, "syncService", "s", true, "syncService")
	err := cobra.MarkFlagRequired(cmd.Flags(), "id")
	if err != nil {
		return err
	}
	err = cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return err
	}
	err = cobra.MarkFlagRequired(cmd.Flags(), "number")
	if err != nil {
		return err
	}
	root.AddCommand(cmd)
	return nil
}
