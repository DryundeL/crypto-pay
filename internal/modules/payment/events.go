package payment

import "time"

const (
	EventPaymentDetected  = "payment.detected"
	EventPaymentConfirmed = "payment.confirmed"
)

type PaymentDetected struct {
	PaymentID  string    `json:"payment_id"`
	InvoiceID  string    `json:"invoice_id"`
	TxHash     string    `json:"tx_hash"`
	Network    string    `json:"network"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (PaymentDetected) EventName() string { return EventPaymentDetected }

type PaymentConfirmed struct {
	PaymentID  string    `json:"payment_id"`
	InvoiceID  string    `json:"invoice_id"`
	TxHash     string    `json:"tx_hash"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (PaymentConfirmed) EventName() string { return EventPaymentConfirmed }
