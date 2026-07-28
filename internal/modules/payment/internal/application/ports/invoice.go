package ports

import (
	"context"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice"
)

// InvoiceLookup finds pending invoices by deposit address.
type InvoiceLookup interface {
	FindPendingByAddress(ctx context.Context, network, address string) (invoice.InvoiceRef, error)
}

// InvoiceLifecycle drives invoice status transitions after payment progress.
type InvoiceLifecycle interface {
	MarkConfirming(ctx context.Context, invoiceID string) error
	MarkPaid(ctx context.Context, invoiceID, txHash string) error
}
