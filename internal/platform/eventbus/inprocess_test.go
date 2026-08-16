package eventbus

import (
	"context"
	"errors"
	"testing"
)

func TestInProcessDeliversToTopicSubscribers(t *testing.T) {
	bus := NewInProcess()
	var paid, other int
	bus.Subscribe("invoice.paid", func(context.Context, string, []byte) error {
		paid++
		return nil
	})
	bus.Subscribe("invoice.expired", func(context.Context, string, []byte) error {
		other++
		return nil
	})

	if err := bus.Publish(context.Background(), "invoice.paid", []byte(`{"ok":true}`)); err != nil {
		t.Fatal(err)
	}
	if paid != 1 || other != 0 {
		t.Fatalf("paid=%d other=%d, want paid=1 other=0", paid, other)
	}

	if err := bus.Publish(context.Background(), "unknown.event", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
}

func TestInProcessReturnsHandlerError(t *testing.T) {
	bus := NewInProcess()
	bus.Subscribe("invoice.paid", func(context.Context, string, []byte) error {
		return errors.New("boom")
	})
	if err := bus.Publish(context.Background(), "invoice.paid", nil); err == nil {
		t.Fatal("expected handler error")
	}
}
