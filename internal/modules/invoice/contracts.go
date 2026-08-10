package invoice

import (
	"context"
	"errors"
	"time"
)

// Public contracts of the Invoice bounded context.
// Other modules depend only on these types/interfaces, never on invoice/internal.

// Status is the public invoice lifecycle status.
type Status string

const (
	StatusPending    Status = "pending"
	StatusConfirming Status = "confirming"
	StatusPaid       Status = "paid"
	StatusExpired    Status = "expired"
	StatusCancelled  Status = "cancelled"
)

// ErrNotFound is returned by public Module facades when an invoice cannot be found.
var ErrNotFound = errors.New("invoice not found")

// InvoiceRef is a cross-module read projection for matching deposits to invoices.
type InvoiceRef struct {
	ID         string
	MerchantID string
	Amount     string
	Currency   string
	Network    string
	Address    string
	Status     string
}

// InvoicePaidNotification is a composition-root side-effect payload after MarkPaid.
type InvoicePaidNotification struct {
	InvoiceID  string
	MerchantID string
	Amount     string
	Currency   string
	TxHash     string
	OccurredAt time.Time
}

// PaidNotifier is invoked by Module.MarkPaid after a successful transition.
type PaidNotifier interface {
	NotifyInvoicePaid(ctx context.Context, in InvoicePaidNotification) error
}

// ExpireDueResult is returned by Module.ExpireDue.
type ExpireDueResult struct {
	Expired int
	Skipped int
}
