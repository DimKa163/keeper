package cli

import (
	"fmt"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/spf13/cobra"
)

type CreateBankCardCommandBuilder struct {
	users      *UserService
	data       *DataService
	sync       *SyncService
	key        string
	name       string
	cardNumber string
	expiry     string
	cvv        string
	holderName string
	bankName   string
	cardType   string
	currency   string
	isPrimary  bool
	needSync   bool
}

func NewCreateBankCardCommandBuilder(users *UserService, data *DataService, sync *SyncService) *CreateBankCardCommandBuilder {
	return &CreateBankCardCommandBuilder{
		users: users,
		data:  data,
		sync:  sync,
	}
}

func (bcc *CreateBankCardCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "create-bank-card",
		Short: "Create a bank card",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := bcc.users.Auth(ctx, bcc.key)
			if err != nil {
				return err
			}
			ctx = SetMasterKey(ctx, masterKey)
			id, err := bcc.data.CreateRecord(ctx, &RecordRequest{
				Type:       core.BankCardType,
				Name:       bcc.name,
				CardNumber: bcc.cardNumber,
				Expiry:     bcc.expiry,
				CVV:        bcc.cvv,
				HolderName: bcc.holderName,
				BankName:   bcc.bankName,
				CardType:   bcc.cardType,
				Currency:   bcc.currency,
				IsPrimary:  bcc.isPrimary,
			})
			if err != nil {
				return err
			}
			fmt.Printf("created bank card data: %s\n", id)
			if bcc.needSync && bcc.sync != nil {
				if err = bcc.sync.Sync(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&bcc.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&bcc.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&bcc.cardNumber, "number", "u", "", "card number")
	cmd.Flags().StringVarP(&bcc.expiry, "expire", "e", "", "expire date")
	cmd.Flags().StringVarP(&bcc.cvv, "cvv", "v", "", "CVV")
	cmd.Flags().StringVarP(&bcc.holderName, "holder-name", "o", "", "holder name")
	cmd.Flags().StringVarP(&bcc.bankName, "bank-name", "b", "", "bank name")
	cmd.Flags().StringVarP(&bcc.cardType, "card-type", "t", "", "card type")
	cmd.Flags().StringVarP(&bcc.currency, "currency", "c", "", "currency")
	cmd.Flags().BoolVarP(&bcc.isPrimary, "primary", "p", false, "primary")
	cmd.Flags().BoolVarP(&bcc.needSync, "sync", "s", true, "sync")
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
	users      *UserService
	data       *DataService
	sync       *SyncService
	id         string
	key        string
	name       string
	cardNumber string
	expiry     string
	cvv        string
	holderName string
	bankName   string
	cardType   string
	currency   string
	isPrimary  bool
	needSync   bool
}

func NewUpdateBankCardCommandBuilder(users *UserService, data *DataService, sync *SyncService) *UpdateBankCardCommandBuilder {
	return &UpdateBankCardCommandBuilder{
		users: users,
		data:  data,
		sync:  sync,
	}
}

func (bcc *UpdateBankCardCommandBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "update-bank-card",
		Short: "update a bank card data",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := bcc.users.Auth(ctx, bcc.key)
			if err != nil {
				return err
			}
			ctx = SetMasterKey(ctx, masterKey)
			id, err := bcc.data.UpdateRecord(ctx, bcc.id, &RecordRequest{
				Type:       core.BankCardType,
				Name:       bcc.name,
				CardNumber: bcc.cardNumber,
				Expiry:     bcc.expiry,
				CVV:        bcc.cvv,
				HolderName: bcc.holderName,
				BankName:   bcc.bankName,
				CardType:   bcc.cardType,
				Currency:   bcc.currency,
				IsPrimary:  bcc.isPrimary,
			})
			if err != nil {
				return err
			}
			fmt.Printf("updated bank card data: %s\n", id)
			if bcc.needSync && bcc.sync != nil {
				if err = bcc.sync.Sync(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&bcc.id, "id", "i", "", "identifier")
	cmd.Flags().StringVarP(&bcc.key, "key", "k", "", "key")
	cmd.Flags().StringVarP(&bcc.name, "name", "n", "", "name")
	cmd.Flags().StringVarP(&bcc.cardNumber, "number", "u", "", "card number")
	cmd.Flags().StringVarP(&bcc.expiry, "expire", "e", "", "expire date")
	cmd.Flags().StringVarP(&bcc.cvv, "cvv", "v", "", "CVV")
	cmd.Flags().StringVarP(&bcc.holderName, "holder-name", "o", "", "holder name")
	cmd.Flags().StringVarP(&bcc.bankName, "bank-name", "b", "", "bank name")
	cmd.Flags().StringVarP(&bcc.cardType, "card-type", "t", "", "card type")
	cmd.Flags().StringVarP(&bcc.currency, "currency", "c", "", "currency")
	cmd.Flags().BoolVarP(&bcc.isPrimary, "primary", "p", false, "primary")
	cmd.Flags().BoolVarP(&bcc.needSync, "sync", "s", true, "sync")
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
