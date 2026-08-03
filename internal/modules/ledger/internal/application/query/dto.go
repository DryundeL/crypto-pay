package query

import "time"

type BalanceDTO struct {
	AccountID string `json:"account_id"`
	Currency  string `json:"currency"`
	Kind      string `json:"kind"`
	Balance   string `json:"balance"`
}

type JournalDTO struct {
	ID            string    `json:"id"`
	MerchantID    string    `json:"merchant_id"`
	ReferenceType string    `json:"reference_type"`
	ReferenceID   string    `json:"reference_id"`
	Amount        string    `json:"amount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}
