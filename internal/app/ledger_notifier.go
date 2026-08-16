package app

import (
	"context"
	"fmt"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger"
)

type journalCrediter interface {
	PostJournal(ctx context.Context, in ledger.PostJournalInput) (ledger.PostJournalResult, error)
}

// ledgerPaidNotifier credits the merchant ledger after invoice.paid (sync, idempotent).
// Webhooks are enqueued by outbox consumers in the worker, not here.
type ledgerPaidNotifier struct {
	ledger journalCrediter
}

// NewLedgerPaidNotifier wires invoice.paid → ledger.PostJournal for API and scanner.
func NewLedgerPaidNotifier(ledgerModule *ledger.Module) ledgerPaidNotifier {
	return ledgerPaidNotifier{ledger: ledgerModule}
}

func (n ledgerPaidNotifier) NotifyInvoicePaid(ctx context.Context, in invoice.InvoicePaidNotification) error {
	_, err := n.ledger.PostJournal(ctx, ledger.PostJournalInput{
		IdempotencyKey: fmt.Sprintf("invoice_paid:%s", in.InvoiceID),
		MerchantID:     in.MerchantID,
		Amount:         in.Amount,
		Currency:       in.Currency,
		ReferenceType:  "invoice",
		ReferenceID:    in.InvoiceID,
	})
	return err
}
