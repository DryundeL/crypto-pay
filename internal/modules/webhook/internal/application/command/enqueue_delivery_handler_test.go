package command

import (
	"context"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/domain"
)

type fakeRepo struct {
	byKey map[string]*domain.Delivery
	saved []*domain.Delivery
}

func (r *fakeRepo) Save(ctx context.Context, d *domain.Delivery) error {
	if r.byKey == nil {
		r.byKey = map[string]*domain.Delivery{}
	}
	r.byKey[d.IdempotencyKey()] = d
	r.saved = append(r.saved, d)
	return nil
}

func (r *fakeRepo) FindByID(ctx context.Context, id domain.DeliveryID) (*domain.Delivery, error) {
	return nil, domain.ErrDeliveryNotFound
}

func (r *fakeRepo) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Delivery, error) {
	if d, ok := r.byKey[key]; ok {
		return d, nil
	}
	return nil, domain.ErrDeliveryNotFound
}

func (r *fakeRepo) FindDue(ctx context.Context, now time.Time, limit int) ([]*domain.Delivery, error) {
	return nil, nil
}

type fakeTx struct{}

func (fakeTx) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fakeMerchant struct{ url string }

func (m fakeMerchant) WebhookURL(ctx context.Context, merchantID string) (string, error) {
	return m.url, nil
}

func (m fakeMerchant) WebhookSecret(context.Context, string) (string, error) {
	return "whsec_test-secret-at-least-32-chars", nil
}

func TestEnqueueDeliveryIdempotent(t *testing.T) {
	repo := &fakeRepo{}
	h := NewEnqueueDeliveryHandler(repo, fakeTx{}, fakeMerchant{url: "https://example.com/hook"})
	h.newID = func() string { return "fixed-id" }
	h.now = func() time.Time { return time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC) }

	cmd := EnqueueDelivery{
		MerchantID: "m-1",
		EventName:  "invoice.paid",
		SourceID:   "inv-1",
		Data:       map[string]any{"invoice_id": "inv-1"},
	}
	first, err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Created || first.ID != "fixed-id" {
		t.Fatalf("first = %+v", first)
	}

	second, err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if second.Created || second.ID != first.ID {
		t.Fatalf("second = %+v", second)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("saved %d times", len(repo.saved))
	}
}

func TestEnqueueDeliverySkipsEmptyURL(t *testing.T) {
	h := NewEnqueueDeliveryHandler(&fakeRepo{}, fakeTx{}, fakeMerchant{url: ""})
	result, err := h.Handle(context.Background(), EnqueueDelivery{
		MerchantID: "m-1",
		EventName:  "invoice.paid",
		SourceID:   "inv-1",
		Data:       map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Skipped {
		t.Fatalf("result = %+v", result)
	}
}
