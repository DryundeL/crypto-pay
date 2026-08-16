package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal"
	"github.com/DryundeL/crypto-pay/internal/platform/eventbus"
)

type deliveryEnqueuer interface {
	Enqueue(ctx context.Context, in webhook.EnqueueInput) (webhook.EnqueueResult, error)
}

type subscriber interface {
	Subscribe(topic string, h eventbus.Handler)
}

// RegisterWebhookConsumers maps integration events to webhook.Enqueue.
// At-least-once: Enqueue is idempotent on event_name:source_id.
func RegisterWebhookConsumers(bus subscriber, webhooks deliveryEnqueuer) {
	if bus == nil {
		panic("app: event bus is required")
	}
	if webhooks == nil {
		panic("app: webhook module is required")
	}
	bus.Subscribe(invoice.EventInvoicePaid, func(ctx context.Context, _ string, payload []byte) error {
		return enqueueInvoicePaid(ctx, webhooks, payload)
	})
	bus.Subscribe(withdrawal.EventWithdrawalCompleted, func(ctx context.Context, _ string, payload []byte) error {
		return enqueueWithdrawalCompleted(ctx, webhooks, payload)
	})
}

func enqueueInvoicePaid(ctx context.Context, webhooks deliveryEnqueuer, payload []byte) error {
	var ev invoice.InvoicePaid
	if err := json.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("decode invoice.paid: %w", err)
	}
	if ev.InvoiceID == "" || ev.MerchantID == "" {
		return fmt.Errorf("invoice.paid: invoice_id and merchant_id are required")
	}
	_, err := webhooks.Enqueue(ctx, webhook.EnqueueInput{
		MerchantID: ev.MerchantID,
		EventName:  invoice.EventInvoicePaid,
		SourceID:   ev.InvoiceID,
		Data: map[string]any{
			"invoice_id":  ev.InvoiceID,
			"merchant_id": ev.MerchantID,
			"amount":      ev.Amount,
			"currency":    ev.Currency,
			"tx_hash":     ev.TxHash,
		},
		OccurredAt: ev.OccurredAt,
	})
	if err != nil {
		return fmt.Errorf("enqueue invoice.paid: %w", err)
	}
	return nil
}

func enqueueWithdrawalCompleted(ctx context.Context, webhooks deliveryEnqueuer, payload []byte) error {
	var ev withdrawal.WithdrawalCompleted
	if err := json.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("decode withdrawal.completed: %w", err)
	}
	if ev.WithdrawalID == "" || ev.MerchantID == "" {
		return fmt.Errorf("withdrawal.completed: withdrawal_id and merchant_id are required")
	}
	_, err := webhooks.Enqueue(ctx, webhook.EnqueueInput{
		MerchantID: ev.MerchantID,
		EventName:  withdrawal.EventWithdrawalCompleted,
		SourceID:   ev.WithdrawalID,
		Data: map[string]any{
			"withdrawal_id": ev.WithdrawalID,
			"merchant_id":   ev.MerchantID,
			"tx_hash":       ev.TxHash,
			"amount":        ev.Amount,
			"currency":      ev.Currency,
			"network":       ev.Network,
		},
		OccurredAt: ev.OccurredAt,
	})
	if err != nil {
		return fmt.Errorf("enqueue withdrawal.completed: %w", err)
	}
	return nil
}
