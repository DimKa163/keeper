package commands

import (
	"encoding/json"
	"errors"

	"github.com/DimKa163/keeper/internal/cli/core"
)

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
		return mapBankCard(record.ID, &bankCard)
	case core.OtherType:
		var binary core.Binary
		if err := json.Unmarshal(data, &binary); err != nil {
			return "", err
		}
		return mapOther(record.ID, &binary)
	}
	return "", errors.New("invalid record")
}

func mapLoginPass(id string, loginPass *core.LoginPass) (string, error) {
	item := struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Login string `json:"login"`
		Pass  string `json:"pass"`
		URL   string `json:"url"`
	}{
		ID:    id,
		Name:  loginPass.Name,
		Login: loginPass.Login,
		Pass:  loginPass.Pass,
		URL:   loginPass.URL,
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

func mapBankCard(id string, bankCard *core.BankCard) (string, error) {
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
