package command_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/eventbus"
)

type memInvoices struct {
	mu   sync.Mutex
	byID map[string]*domain.Invoice
}

func (m *memInvoices) Save(_ context.Context, inv *domain.Invoice) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byID == nil {
		m.byID = map[string]*domain.Invoice{}
	}
	m.byID[inv.ID().String()] = inv
	return nil
}

func (m *memInvoices) FindByID(_ context.Context, id domain.InvoiceID) (*domain.Invoice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	inv, ok := m.byID[id.String()]
	if !ok {
		return nil, domain.ErrInvoiceNotFound
	}
	return inv, nil
}

func (m *memInvoices) ListExpiredPendingIDs(_ context.Context, now time.Time, limit int) ([]domain.InvoiceID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.InvoiceID
	for _, inv := range m.byID {
		if inv.Status() != domain.StatusPending {
			continue
		}
		if !inv.ExpiresAt().Before(now) {
			continue
		}
		out = append(out, inv.ID())
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

type noopTX struct{}

func (noopTX) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type memPub struct {
	events []eventbus.IntegrationEvent
}

func (p *memPub) Publish(_ context.Context, _, _ string, event eventbus.IntegrationEvent) error {
	p.events = append(p.events, event)
	return nil
}

func TestExpireDueExpiresPendingPastTTL(t *testing.T) {
	money, err := domain.NewMoney("1.00", "ETH")
	if err != nil {
		t.Fatal(err)
	}

	createdAt := time.Now().UTC().Add(-2 * time.Hour)
	due, err := domain.Create(domain.CreateParams{
		ID:         domain.InvoiceID("inv-due"),
		MerchantID: "m-1",
		Amount:     money,
		Network:    "evm:sepolia",
		Address:    "0xabc",
		ExpiresAt:  time.Now().UTC().Add(-time.Minute),
		Now:        createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	alive, err := domain.Create(domain.CreateParams{
		ID:         domain.InvoiceID("inv-alive"),
		MerchantID: "m-1",
		Amount:     money,
		Network:    "evm:sepolia",
		Address:    "0xdef",
		ExpiresAt:  time.Now().UTC().Add(time.Hour),
		Now:        time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}

	repo := &memInvoices{byID: map[string]*domain.Invoice{
		"inv-due":   due,
		"inv-alive": alive,
	}}
	pub := &memPub{}
	h := command.NewExpireDueHandler(repo, command.NewExpireInvoiceHandler(repo, noopTX{}, pub))

	result, err := h.Handle(context.Background(), command.ExpireDue{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if result.Expired != 1 {
		t.Fatalf("Expired = %d, want 1", result.Expired)
	}
	if due.Status() != domain.StatusExpired {
		t.Fatalf("due status = %s, want expired", due.Status())
	}
	if alive.Status() != domain.StatusPending {
		t.Fatalf("alive status = %s, want pending", alive.Status())
	}
	if len(pub.events) != 1 || pub.events[0].EventName() != "invoice.expired" {
		t.Fatalf("events = %#v, want invoice.expired", pub.events)
	}
}
