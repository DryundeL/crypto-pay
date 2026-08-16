package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/platform/eventbus"
)

type memPending struct {
	mu   sync.Mutex
	msgs []Message
}

func (m *memPending) ProcessPending(ctx context.Context, limit int, fn func(context.Context, Message) error) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	n := 0
	for i := range m.msgs {
		if n >= limit {
			break
		}
		if m.msgs[i].Status != StatusPending {
			continue
		}
		if err := fn(ctx, m.msgs[i]); err != nil {
			return n, err
		}
		now := time.Now().UTC()
		m.msgs[i].Status = StatusSent
		m.msgs[i].PublishedAt = &now
		n++
	}
	return n, nil
}

func TestRelayPendingPublishesAndMarksSent(t *testing.T) {
	src := &memPending{msgs: []Message{
		{
			ID:        "1",
			EventName: "invoice.paid",
			Payload:   json.RawMessage(`{"invoice_id":"inv-1"}`),
			Status:    StatusPending,
		},
		{
			ID:        "2",
			EventName: "merchant.created",
			Payload:   json.RawMessage(`{"merchant_id":"m-1"}`),
			Status:    StatusSent,
		},
		{
			ID:        "3",
			EventName: "invoice.paid",
			Payload:   json.RawMessage(`{"invoice_id":"inv-2"}`),
			Status:    StatusPending,
		},
	}}

	bus := eventbus.NewInProcess()
	var seen []string
	bus.Subscribe("invoice.paid", func(_ context.Context, topic string, payload []byte) error {
		seen = append(seen, topic+":"+string(payload))
		return nil
	})

	r := &relayService{proc: src, pub: bus}
	n, err := r.RelayPending(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("published = %d, want 2", n)
	}
	if len(seen) != 2 {
		t.Fatalf("consumer saw %d events, want 2: %v", len(seen), seen)
	}
	if src.msgs[0].Status != StatusSent || src.msgs[0].PublishedAt == nil {
		t.Fatalf("msg 1 status = %s, published_at nil=%v", src.msgs[0].Status, src.msgs[0].PublishedAt == nil)
	}
	if src.msgs[1].Status != StatusSent {
		t.Fatalf("already-sent msg changed: %s", src.msgs[1].Status)
	}
	if src.msgs[2].Status != StatusSent {
		t.Fatalf("msg 3 status = %s, want sent", src.msgs[2].Status)
	}
}

func TestRelayPendingKeepsPendingOnPublishError(t *testing.T) {
	src := &memPending{msgs: []Message{{
		ID:        "1",
		EventName: "invoice.paid",
		Payload:   json.RawMessage(`{}`),
		Status:    StatusPending,
	}}}

	bus := eventbus.NewInProcess()
	bus.Subscribe("invoice.paid", func(context.Context, string, []byte) error {
		return errors.New("consumer down")
	})

	r := &relayService{proc: src, pub: bus}
	n, err := r.RelayPending(context.Background(), 10)
	if err == nil {
		t.Fatal("expected publish error")
	}
	if n != 0 {
		t.Fatalf("published = %d, want 0", n)
	}
	if src.msgs[0].Status != StatusPending {
		t.Fatalf("status = %s, want pending", src.msgs[0].Status)
	}
}
