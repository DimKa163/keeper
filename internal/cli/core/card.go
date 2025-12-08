package core

type BankCard struct {
	Name       string `json:"name"`
	CardNumber string `json:"card_number"`
	Expiry     string `json:"expiry"`
	CVV        string `json:"cvv"`
	HolderName string `json:"holder_name"`
	BankName   string `json:"bank_name,omitempty"`
	CardType   string `json:"card_type,omitempty"`
	Currency   string `json:"currency,omitempty"`
	IsPrimary  bool   `json:"is_primary"`
}
