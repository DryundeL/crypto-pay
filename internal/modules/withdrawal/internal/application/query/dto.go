package query

import "time"

type WithdrawalDTO struct {
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
