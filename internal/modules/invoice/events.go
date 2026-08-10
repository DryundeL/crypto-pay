package invoice

import "time"

// Integration events of the Invoice bounded context.
// Persisted via transactional outbox; consumed by other modules asynchronously.

const (
	EventInvoiceCreated = "invoice.created"
	EventInvoicePaid    = "invoice.paid"
	EventInvoiceExpired = "invoice.expired"
)

type InvoiceCreated struct {
	InvoiceID  string    `json:"invoice_id"`
	MerchantID string    `json:"merchant_id"`
	Amount     string    `json:"amount"`
	Currency   string    `json:"currency"`
	Network    string    `json:"network"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (InvoiceCreated) EventName() string { return EventInvoiceCreated }

type InvoicePaid struct {
	InvoiceID  string    `json:"invoice_id"`
	MerchantID string    `json:"merchant_id"`
	Amount     string    `json:"amount"`
	Currency   string    `json:"currency"`
	TxHash     string    `json:"tx_hash"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (InvoicePaid) EventName() string { return EventInvoicePaid }

type InvoiceExpired struct {
	InvoiceID  string    `json:"invoice_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (InvoiceExpired) EventName() string { return EventInvoiceExpired }
