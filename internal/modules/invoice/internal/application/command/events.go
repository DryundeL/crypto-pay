package command

import "time"

// Integration event payloads for outbox (mirror public invoice.events contracts).
// Kept here to avoid import cycle: invoice -> command -> invoice.

type invoiceCreatedEvent struct {
	InvoiceID  string    `json:"invoice_id"`
	MerchantID string    `json:"merchant_id"`
	Amount     string    `json:"amount"`
	Currency   string    `json:"currency"`
	Network    string    `json:"network"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (invoiceCreatedEvent) EventName() string { return "invoice.created" }

type invoicePaidEvent struct {
	InvoiceID  string    `json:"invoice_id"`
	MerchantID string    `json:"merchant_id"`
	Amount     string    `json:"amount"`
	Currency   string    `json:"currency"`
	TxHash     string    `json:"tx_hash"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (invoicePaidEvent) EventName() string { return "invoice.paid" }

type invoiceExpiredEvent struct {
	InvoiceID  string    `json:"invoice_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (invoiceExpiredEvent) EventName() string { return "invoice.expired" }
