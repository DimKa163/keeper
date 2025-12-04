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
)

type DataListBuilder struct {
	userService *app.UserService
	dataManager *app.DataManager
	decoder     core.Decoder
	key         string
	limit       int32
	offset      int32
}

func NewDataListBuilder(userService *app.UserService, dataManager *app.DataManager, decoder core.Decoder) *DataListBuilder {
	return &DataListBuilder{userService: userService, dataManager: dataManager, decoder: decoder}
}

func (c *DataListBuilder) Build() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all dataManager",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			masterKey, err := c.userService.Auth(ctx, c.key)
			if err != nil {
				return err
			}
			ctx = common.SetMasterKey(ctx, masterKey)
			records, err := c.dataManager.GetAll(ctx, c.limit, c.offset)
			if err != nil {
				return err
			}
			var js string
			for _, record := range records {
				js, err = c.mapRecord(ctx, record)
				if err != nil {
					return err
				}
				fmt.Println(js)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&c.key, "key", "k", "", "key")
	cmd.Flags().Int32VarP(&c.limit, "limit", "l", 5, "limit")
	cmd.Flags().Int32VarP(&c.offset, "offset", "o", 0, "offset")
	err := cobra.MarkFlagRequired(cmd.Flags(), "key")
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (c *DataListBuilder) mapRecord(ctx context.Context, record *core.Record) (string, error) {
	masterKey, err := common.GetMasterKey(ctx)
	if err != nil {
		return "", err
	}
	switch record.Type {
	case core.LoginPassType:
		data, err := record.DecodeLoginPass(c.decoder, masterKey)
		if err != nil {
			return "", err
		}
		return mapLoginPass(record.ID, data)
	case core.TextType:
		data, err := record.DecodeText(c.decoder, masterKey)
		if err != nil {
			return "", err
		}
		return mapText(record.ID, data)
	case core.BankCardType:
		data, err := record.DecodeBankCard(c.decoder, masterKey)
		if err != nil {
			return "", err
		}
		return mapBakCard(record.ID, data)
	case core.OtherType:
		data, err := record.DecodeBinary(c.decoder, masterKey)
		if err != nil {
			return "", err
		}
		return mapOther(record.ID, data)
	}
	return "", errors.New("invalid record")
}

func mapLoginPass(id string, loginPass *core.LoginPass) (string, error) {
	item := struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Login string `json:"login"`
		Pass  string `json:"pass"`
		Url   string `json:"url"`
	}{
		ID:    id,
		Name:  loginPass.Name,
		Login: loginPass.Login,
		Pass:  loginPass.Pass,
		Url:   loginPass.Url,
	}
	jsonData, err := json.MarshalIndent(item, "", " ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func mapText(id string, text *core.Text) (string, error) {
	item := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Text string `json:"text"`
	}{
		ID:   id,
		Name: text.Name,
		Text: text.Content,
	}
	jsonData, err := json.MarshalIndent(item, "", " ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func mapBakCard(id string, bankCard *core.BankCard) (string, error) {
	item := struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		CardNumber string `json:"card_number"`
		Expiry     string `json:"expiry"`
		CVV        string `json:"cvv"`
		HolderName string `json:"holder_name"`
		BankName   string `json:"bank_name,omitempty"`
		CardType   string `json:"card_type,omitempty"`
		Currency   string `json:"currency,omitempty"`
		IsPrimary  bool   `json:"is_primary"`
	}{
		ID:         id,
		Name:       bankCard.Name,
		CardNumber: bankCard.CardNumber,
		Expiry:     bankCard.Expiry,
		CVV:        bankCard.CVV,
		HolderName: bankCard.HolderName,
		BankName:   bankCard.BankName,
		CardType:   bankCard.CardType,
		Currency:   bankCard.Currency,
		IsPrimary:  bankCard.IsPrimary,
	}
	jsonData, err := json.MarshalIndent(item, "", " ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func mapOther(id string, other *core.Binary) (string, error) {
	item := struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		MIMEType  string `json:"mime_type"`
		SizeBytes int64  `json:"size"`
	}{
		ID:        id,
		Name:      other.Name,
		MIMEType:  other.MIMEType,
		SizeBytes: other.SizeBytes,
	}
	jsonData, err := json.MarshalIndent(item, "", " ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
