package http

import "time"

type createWithdrawalRequest struct {
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`
	Network        string `json:"network"`
	ToAddress      string `json:"to_address"`
	IdempotencyKey string `json:"idempotency_key"`
}

type withdrawalResponse struct {
	ID             string    `json:"id"`
	MerchantID     string    `json:"merchant_id"`
	IdempotencyKey string    `json:"idempotency_key"`
	Amount         string    `json:"amount"`
	Currency       string    `json:"currency"`
	Network        string    `json:"network"`
	ToAddress      string    `json:"to_address"`
	Status         string    `json:"status"`
	TxHash         string    `json:"tx_hash,omitempty"`
	JournalID      string    `json:"journal_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}
