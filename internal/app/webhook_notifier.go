package app

import (
	"context"
	"fmt"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal"
)

type journalCrediter interface {
	PostJournal(ctx context.Context, in ledger.PostJournalInput) (ledger.PostJournalResult, error)
}

type deliveryEnqueuer interface {
	Enqueue(ctx context.Context, in webhook.EnqueueInput) (webhook.EnqueueResult, error)
}

// sideEffectNotifier bridges invoice/withdrawal facades to ledger credit + webhook.Enqueue.
// Sync path for v1; phase 2 moves these to outbox consumers.
type sideEffectNotifier struct {
	ledger   journalCrediter
	webhooks deliveryEnqueuer
}

func (n sideEffectNotifier) NotifyInvoicePaid(ctx context.Context, in invoice.InvoicePaidNotification) error {
	if _, err := n.ledger.PostJournal(ctx, ledger.PostJournalInput{
		IdempotencyKey: fmt.Sprintf("invoice_paid:%s", in.InvoiceID),
		MerchantID:     in.MerchantID,
		Amount:         in.Amount,
		Currency:       in.Currency,
		ReferenceType:  "invoice",
		ReferenceID:    in.InvoiceID,
	}); err != nil {
		return err
	}

	_, err := n.webhooks.Enqueue(ctx, webhook.EnqueueInput{
		MerchantID: in.MerchantID,
		EventName:  invoice.EventInvoicePaid,
		SourceID:   in.InvoiceID,
		Data: map[string]any{
			"invoice_id":  in.InvoiceID,
			"merchant_id": in.MerchantID,
			"amount":      in.Amount,
			"currency":    in.Currency,
			"tx_hash":     in.TxHash,
		},
		OccurredAt: in.OccurredAt,
	})
	return err
}

func (n sideEffectNotifier) NotifyWithdrawalCompleted(ctx context.Context, in withdrawal.WithdrawalCompletedNotification) error {
	_, err := n.webhooks.Enqueue(ctx, webhook.EnqueueInput{
		MerchantID: in.MerchantID,
		EventName:  withdrawal.EventWithdrawalCompleted,
		SourceID:   in.WithdrawalID,
		Data: map[string]any{
			"withdrawal_id": in.WithdrawalID,
			"merchant_id":   in.MerchantID,
			"tx_hash":       in.TxHash,
			"amount":        in.Amount,
			"currency":      in.Currency,
			"network":       in.Network,
		},
		OccurredAt: in.OccurredAt,
	})
	return err
}
