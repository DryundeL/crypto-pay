package payment

import (
	"errors"
	"time"
)

// Public contracts of the Payment bounded context.
type Status string

const (
	StatusDetected   Status = "detected"
	StatusConfirming Status = "confirming"
	StatusConfirmed  Status = "confirmed"
	StatusFailed     Status = "failed"
)

// ErrNotFound is returned by public Module facades when a payment cannot be found.
var ErrNotFound = errors.New("payment not found")

// ErrInvalid is returned for validation / business-rule failures on public facades.
var ErrInvalid = errors.New("invalid payment")

// RecordObservedInput is the public input for recording a chain observation.
type RecordObservedInput struct {
	MerchantID    string
	Network       string
	TxHash        string
	ToAddress     string
	Amount        string
	Currency      string
	Confirmations int
}

// ConfirmInput is the public input for confirming a payment by network+tx.
type ConfirmInput struct {
	MerchantID string
	Network    string
	TxHash     string
}

// PaymentView is a public projection of a payment for cross-module callers.
type PaymentView struct {
	ID            string
	InvoiceID     string
	MerchantID    string
	Status        string
	Network       string
	TxHash        string
	ToAddress     string
	Amount        string
	Currency      string
	Confirmations int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
