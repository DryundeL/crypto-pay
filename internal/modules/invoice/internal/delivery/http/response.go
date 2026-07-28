package http

import "time"

type createInvoiceRequest struct {
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	Network          string `json:"network"`
	ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
}

type invoiceResponse struct {
	ID         string    `json:"id"`
	MerchantID string    `json:"merchant_id"`
	Status     string    `json:"status"`
	Amount     string    `json:"amount"`
	Currency   string    `json:"currency"`
	Network    string    `json:"network"`
	Address    string    `json:"address"`
	TxHash     string    `json:"tx_hash,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}
