package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook"
)

type stubLedger struct {
	calls []ledger.PostJournalInput
	err   error
}

func (s *stubLedger) PostJournal(_ context.Context, in ledger.PostJournalInput) (ledger.PostJournalResult, error) {
	s.calls = append(s.calls, in)
	if s.err != nil {
		return ledger.PostJournalResult{}, s.err
	}
	return ledger.PostJournalResult{JournalID: "j-1", Created: len(s.calls) == 1}, nil
}

type stubWebhook struct {
	calls int
}

func (s *stubWebhook) Enqueue(_ context.Context, _ webhook.EnqueueInput) (webhook.EnqueueResult, error) {
	s.calls++
	return webhook.EnqueueResult{ID: "d-1", Created: true}, nil
}

func TestNotifyInvoicePaidCreditsLedger(t *testing.T) {
	led := &stubLedger{}
	wh := &stubWebhook{}
	n := sideEffectNotifier{ledger: led, webhooks: wh}

	in := invoice.InvoicePaidNotification{
		InvoiceID:  "inv-1",
		MerchantID: "m-1",
		Amount:     "1.50",
		Currency:   "ETH",
		TxHash:     "0xabc",
		OccurredAt: time.Unix(1, 0).UTC(),
	}

	if err := n.NotifyInvoicePaid(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	if err := n.NotifyInvoicePaid(context.Background(), in); err != nil {
		t.Fatal(err)
	}

	if len(led.calls) != 2 {
		t.Fatalf("PostJournal calls = %d, want 2", len(led.calls))
	}
	got := led.calls[0]
	if got.IdempotencyKey != "invoice_paid:inv-1" {
		t.Fatalf("IdempotencyKey = %q, want invoice_paid:inv-1", got.IdempotencyKey)
	}
	if got.MerchantID != "m-1" || got.Amount != "1.50" || got.Currency != "ETH" {
		t.Fatalf("unexpected credit input: %+v", got)
	}
	if got.ReferenceType != "invoice" || got.ReferenceID != "inv-1" {
		t.Fatalf("unexpected reference: %+v", got)
	}
	if led.calls[1].IdempotencyKey != led.calls[0].IdempotencyKey {
		t.Fatal("retry must reuse the same idempotency key")
	}
	if wh.calls != 2 {
		t.Fatalf("webhook Enqueue calls = %d, want 2", wh.calls)
	}
}

func TestNotifyInvoicePaidStopsOnLedgerError(t *testing.T) {
	led := &stubLedger{err: errors.New("ledger down")}
	wh := &stubWebhook{}
	n := sideEffectNotifier{ledger: led, webhooks: wh}

	err := n.NotifyInvoicePaid(context.Background(), invoice.InvoicePaidNotification{
		InvoiceID:  "inv-1",
		MerchantID: "m-1",
		Amount:     "1",
		Currency:   "ETH",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if wh.calls != 0 {
		t.Fatal("webhook must not run after ledger failure")
	}
}
