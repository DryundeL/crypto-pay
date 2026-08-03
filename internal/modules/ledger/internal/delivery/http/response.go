package http

import "time"

type balanceResponse struct {
	AccountID string `json:"account_id"`
	Currency  string `json:"currency"`
	Kind      string `json:"kind"`
	Balance   string `json:"balance"`
}

type journalResponse struct {
	ID            string    `json:"id"`
	MerchantID    string    `json:"merchant_id"`
	ReferenceType string    `json:"reference_type"`
	ReferenceID   string    `json:"reference_id"`
	Amount        string    `json:"amount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}
