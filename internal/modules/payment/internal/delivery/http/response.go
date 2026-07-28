package http

import "time"

type observedRequest struct {
	Network       string `json:"network"`
	TxHash        string `json:"tx_hash"`
	ToAddress     string `json:"to_address"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Confirmations int    `json:"confirmations"`
}

type confirmedRequest struct {
	Network string `json:"network"`
	TxHash  string `json:"tx_hash"`
}

type paymentResponse struct {
	ID            string    `json:"id"`
	InvoiceID     string    `json:"invoice_id"`
	MerchantID    string    `json:"merchant_id"`
	Status        string    `json:"status"`
	Network       string    `json:"network"`
	TxHash        string    `json:"tx_hash"`
	ToAddress     string    `json:"to_address"`
	Amount        string    `json:"amount"`
	Currency      string    `json:"currency"`
	Confirmations int       `json:"confirmations"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}
