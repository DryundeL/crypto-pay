package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/domain"
)

func TestEnqueueAndMarkSent(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	d, err := domain.Enqueue(domain.EnqueueParams{
		ID:         "d-1",
		MerchantID: "m-1",
		EventName:  "invoice.paid",
		SourceID:   "inv-1",
		URL:        "https://example.com/hook",
		Payload:    []byte(`{"ok":true}`),
		Now:        now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.Status() != domain.StatusPending {
		t.Fatalf("status = %s", d.Status())
	}
	if d.IdempotencyKey() != "invoice.paid:inv-1" {
		t.Fatalf("idempotency = %s", d.IdempotencyKey())
	}

	if err := d.MarkSent(200, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if d.Status() != domain.StatusSent || d.AttemptCount() != 1 || d.DeliveredAt() == nil {
		t.Fatalf("status=%s attempts=%d delivered=%v", d.Status(), d.AttemptCount(), d.DeliveredAt())
	}
}

func TestEnqueueRejectsEmptyPayload(t *testing.T) {
	_, err := domain.Enqueue(domain.EnqueueParams{
		ID:         "d-1",
		MerchantID: "m-1",
		EventName:  "invoice.paid",
		SourceID:   "inv-1",
		URL:        "https://example.com/hook",
		Payload:    nil,
		Now:        time.Now().UTC(),
	})
	if !errors.Is(err, domain.ErrInvalidDelivery) {
		t.Fatalf("err = %v", err)
	}
}

func TestMarkAttemptFailedRetriesThenTerminal(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	d, err := domain.Enqueue(domain.EnqueueParams{
		ID:          "d-1",
		MerchantID:  "m-1",
		EventName:   "invoice.paid",
		SourceID:    "inv-1",
		URL:         "https://example.com/hook",
		Payload:     []byte(`{}`),
		MaxAttempts: 2,
		Now:         now,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := d.MarkAttemptFailed(500, "boom", true, now); err != nil {
		t.Fatal(err)
	}
	if d.Status() != domain.StatusPending || d.AttemptCount() != 1 {
		t.Fatalf("status=%s attempts=%d", d.Status(), d.AttemptCount())
	}
	if !d.NextAttemptAt().After(now) {
		t.Fatalf("expected backoff next_attempt_at, got %s", d.NextAttemptAt())
	}

	if err := d.MarkAttemptFailed(500, "boom", true, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if d.Status() != domain.StatusFailed || d.AttemptCount() != 2 {
		t.Fatalf("status=%s attempts=%d", d.Status(), d.AttemptCount())
	}
}

func TestMarkAttemptFailedNonRetryable(t *testing.T) {
	now := time.Now().UTC()
	d, err := domain.Enqueue(domain.EnqueueParams{
		ID:          "d-1",
		MerchantID:  "m-1",
		EventName:   "invoice.paid",
		SourceID:    "inv-1",
		URL:         "https://example.com/hook",
		Payload:     []byte(`{}`),
		MaxAttempts: 8,
		Now:         now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := d.MarkAttemptFailed(400, "bad request", false, now); err != nil {
		t.Fatal(err)
	}
	if d.Status() != domain.StatusFailed {
		t.Fatalf("status = %s", d.Status())
	}
}
