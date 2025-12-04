package _interface

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/app"
	"github.com/DimKa163/keeper/internal/cli/common"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/spf13/cobra"
)

type CreateBankCardCommandBuilder struct {
	userService *app.UserService
	dataManager *app.DataManager
	key         string
	name        string
	cardNumber  string
	expiry      string
	cvv         string
	holderName  string
	bankName    string
	cardType    string
	currency    string
	isPrimary   bool
	needSync    bool
}

func NewCreateBankCardCommandBuilder(users *app.UserService, dataManager *app.DataManager) *CreateBankCardCommandBuilder {
	return &CreateBankCardCommandBuilder{
		userService: users,
		dataManager: dataManager,
	}
}

func (c *CreateBankCardCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "create-bank-card",
		Short: "Create a bank card",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.userService.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err := c.dataManager.Create(ctx, &app.RecordRequest{
				Type:       core.BankCardType,
				Name:       c.name,
				CardNumber: c.cardNumber,
				Expiry:     c.expiry,
				CVV:        c.cvv,
				HolderName: c.holderName,
				BankName:   c.bankName,
				CardType:   c.cardType,
				Currency:   c.currency,
				IsPrimary:  c.isPrimary,
			}, c.needSync)
			if err != nil {
				return err
			}
			fmt.Printf("created bank card dataManager: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&c.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&c.cardNumber, "number", "u", "", "card number")
	cmd.Flags().StringVarP(&c.expiry, "expire", "e", "", "expire date")
	cmd.Flags().StringVarP(&c.cvv, "cvv", "v", "", "CVV")
	cmd.Flags().StringVarP(&c.holderName, "holder-name", "o", "", "holder name")
	cmd.Flags().StringVarP(&c.bankName, "bank-name", "b", "", "bank name")
	cmd.Flags().StringVarP(&c.cardType, "card-type", "t", "", "card type")
	cmd.Flags().StringVarP(&c.currency, "currency", "c", "", "currency")
	cmd.Flags().BoolVarP(&c.isPrimary, "primary", "p", false, "primary")
	cmd.Flags().BoolVarP(&c.needSync, "syncService", "s", true, "syncService")
	err := cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return nil, err
	}
	err = cobra.MarkFlagRequired(cmd.Flags(), "number")
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

type UpdateBankCardCommandBuilder struct {
	userService *app.UserService
	dataManager *app.DataManager
	id          string
	key         string
	name        string
	cardNumber  string
	expiry      string
	cvv         string
	holderName  string
	bankName    string
	cardType    string
	currency    string
	isPrimary   bool
	needSync    bool
}

func NewUpdateBankCardCommandBuilder(
	userService *app.UserService,
	dataManager *app.DataManager,
) *UpdateBankCardCommandBuilder {
	return &UpdateBankCardCommandBuilder{
		userService: userService,
		dataManager: dataManager,
	}
}

func (c *UpdateBankCardCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "update-bank-card",
		Short: "update a bank card dataManager",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.userService.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			id, err := c.dataManager.Update(ctx, c.id, &app.RecordRequest{
				Type:       core.BankCardType,
				Name:       c.name,
				CardNumber: c.cardNumber,
				Expiry:     c.expiry,
				CVV:        c.cvv,
				HolderName: c.holderName,
				BankName:   c.bankName,
				CardType:   c.cardType,
				Currency:   c.currency,
				IsPrimary:  c.isPrimary,
			}, c.needSync)
			if err != nil {
				return err
			}
			fmt.Printf("updated bank card dataManager: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&c.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&c.cardNumber, "number", "u", "", "card number")
	cmd.Flags().StringVarP(&c.expiry, "expire", "e", "", "expire date")
	cmd.Flags().StringVarP(&c.cvv, "cvv", "v", "", "CVV")
	cmd.Flags().StringVarP(&c.holderName, "holder-name", "o", "", "holder name")
	cmd.Flags().StringVarP(&c.bankName, "bank-name", "b", "", "bank name")
	cmd.Flags().StringVarP(&c.cardType, "card-type", "t", "", "card type")
	cmd.Flags().StringVarP(&c.currency, "currency", "c", "", "currency")
	cmd.Flags().BoolVarP(&c.isPrimary, "primary", "p", false, "primary")
	cmd.Flags().BoolVarP(&c.needSync, "syncService", "s", true, "syncService")
	err := cobra.MarkFlagRequired(cmd.Flags(), "id")
	if err != nil {
		return nil, err
	}
	err = cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return nil, err
	}
	err = cobra.MarkFlagRequired(cmd.Flags(), "number")
	if err != nil {
		return nil, err
	}
	return cmd, nil
}
