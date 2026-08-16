package app

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal"
	"github.com/DryundeL/crypto-pay/internal/platform/eventbus"
)

type stubWebhook struct {
	calls []webhook.EnqueueInput
	err   error
}

func (s *stubWebhook) Enqueue(_ context.Context, in webhook.EnqueueInput) (webhook.EnqueueResult, error) {
	s.calls = append(s.calls, in)
	if s.err != nil {
		return webhook.EnqueueResult{}, s.err
	}
	return webhook.EnqueueResult{ID: "d-1", Created: len(s.calls) == 1}, nil
}

func TestWebhookConsumersEnqueueFromEvents(t *testing.T) {
	bus := eventbus.NewInProcess()
	wh := &stubWebhook{}
	RegisterWebhookConsumers(bus, wh)

	paid := invoice.InvoicePaid{
		InvoiceID:  "inv-1",
		MerchantID: "m-1",
		Amount:     "1.50",
		Currency:   "ETH",
		TxHash:     "0xabc",
		OccurredAt: time.Unix(1, 0).UTC(),
	}
	body, err := json.Marshal(paid)
	if err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(context.Background(), invoice.EventInvoicePaid, body); err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(context.Background(), invoice.EventInvoicePaid, body); err != nil {
		t.Fatal(err)
	}

	completed := withdrawal.WithdrawalCompleted{
		WithdrawalID: "wd-1",
		MerchantID:   "m-1",
		TxHash:       "0xdead",
		Amount:       "2",
		Currency:     "ETH",
		Network:      "evm:sepolia",
		OccurredAt:   time.Unix(2, 0).UTC(),
	}
	wdBody, err := json.Marshal(completed)
	if err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(context.Background(), withdrawal.EventWithdrawalCompleted, wdBody); err != nil {
		t.Fatal(err)
	}

	if len(wh.calls) != 3 {
		t.Fatalf("Enqueue calls = %d, want 3", len(wh.calls))
	}

	got := wh.calls[0]
	if got.EventName != invoice.EventInvoicePaid || got.SourceID != "inv-1" || got.MerchantID != "m-1" {
		t.Fatalf("invoice enqueue = %+v", got)
	}
	if got.Data["amount"] != "1.50" || got.Data["tx_hash"] != "0xabc" {
		t.Fatalf("invoice data = %+v", got.Data)
	}

	wd := wh.calls[2]
	if wd.EventName != withdrawal.EventWithdrawalCompleted || wd.SourceID != "wd-1" {
		t.Fatalf("withdrawal enqueue = %+v", wd)
	}
	if wd.Data["amount"] != "2" || wd.Data["network"] != "evm:sepolia" {
		t.Fatalf("withdrawal data = %+v", wd.Data)
	}
}

func TestWebhookConsumerRejectsInvalidPayload(t *testing.T) {
	bus := eventbus.NewInProcess()
	wh := &stubWebhook{}
	RegisterWebhookConsumers(bus, wh)

	err := bus.Publish(context.Background(), invoice.EventInvoicePaid, []byte(`{`))
	if err == nil {
		t.Fatal("expected decode error")
	}
	if len(wh.calls) != 0 {
		t.Fatal("must not enqueue on decode error")
	}
}
