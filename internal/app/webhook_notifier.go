package app

import (
	"context"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal"
)

// webhookNotifier bridges invoice/withdrawal facades to webhook.Enqueue.
type webhookNotifier struct {
	webhooks *webhook.Module
}

func (n webhookNotifier) NotifyInvoicePaid(ctx context.Context, in invoice.InvoicePaidNotification) error {
	_, err := n.webhooks.Enqueue(ctx, webhook.EnqueueInput{
		MerchantID: in.MerchantID,
		EventName:  invoice.EventInvoicePaid,
		SourceID:   in.InvoiceID,
		Data: map[string]any{
			"invoice_id":  in.InvoiceID,
			"merchant_id": in.MerchantID,
			"tx_hash":     in.TxHash,
		},
		OccurredAt: in.OccurredAt,
	})
	return err
}

func (n webhookNotifier) NotifyWithdrawalCompleted(ctx context.Context, in withdrawal.WithdrawalCompletedNotification) error {
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
