package command

import "time"

type paymentDetectedEvent struct {
	PaymentID  string    `json:"payment_id"`
	InvoiceID  string    `json:"invoice_id"`
	TxHash     string    `json:"tx_hash"`
	Network    string    `json:"network"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (paymentDetectedEvent) EventName() string { return "payment.detected" }

type paymentConfirmedEvent struct {
	PaymentID  string    `json:"payment_id"`
	InvoiceID  string    `json:"invoice_id"`
	TxHash     string    `json:"tx_hash"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (paymentConfirmedEvent) EventName() string { return "payment.confirmed" }
