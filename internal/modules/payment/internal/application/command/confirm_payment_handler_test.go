package command_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/payment"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/eventbus"
)

type memPayments struct {
	mu   sync.Mutex
	byTx map[string]*domain.Payment
}

func (m *memPayments) key(network, txHash string) string { return network + "|" + txHash }

func (m *memPayments) FindByID(context.Context, domain.PaymentID) (*domain.Payment, error) {
	return nil, domain.ErrPaymentNotFound
}

func (m *memPayments) FindByNetworkTxHash(_ context.Context, network, txHash string) (*domain.Payment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.byTx[m.key(network, txHash)]
	if !ok {
		return nil, domain.ErrPaymentNotFound
	}
	return p, nil
}

func (m *memPayments) Save(_ context.Context, payment *domain.Payment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byTx == nil {
		m.byTx = map[string]*domain.Payment{}
	}
	m.byTx[m.key(payment.Network(), payment.TxHash())] = payment
	return nil
}

type noopTX struct{}

func (noopTX) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type memPub struct{}

func (memPub) Publish(context.Context, string, string, eventbus.IntegrationEvent) error { return nil }

type stubInvoice struct {
	paid int
}

func (s *stubInvoice) MarkConfirming(context.Context, string) error { return nil }
func (s *stubInvoice) MarkPaid(context.Context, string, string) error {
	s.paid++
	return nil
}

func TestConfirmPaymentRequiresConfirmations(t *testing.T) {
	now := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	pay, err := domain.Detect(domain.DetectParams{
		ID:            domain.PaymentID("pay-1"),
		InvoiceID:     "inv-1",
		MerchantID:    "m-1",
		Network:       "evm:sepolia",
		TxHash:        "0xabc",
		ToAddress:     "addr-1",
		Amount:        "0.01",
		Currency:      "ETH",
		Confirmations: 0,
		Now:           now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := pay.StartConfirming(now); err != nil {
		t.Fatal(err)
	}

	repo := &memPayments{byTx: map[string]*domain.Payment{
		"evm:sepolia|0xabc": pay,
	}}
	invoice := &stubInvoice{}
	policy := payment.NewConfirmationPolicy(map[string]int{"evm:sepolia": 1})
	h := command.NewConfirmPaymentHandler(repo, noopTX{}, memPub{}, invoice, policy)

	_, err = h.Handle(context.Background(), command.ConfirmPayment{
		MerchantID: "m-1",
		Network:    "evm:sepolia",
		TxHash:     "0xabc",
	})
	if !errors.Is(err, domain.ErrInsufficientConfirmations) {
		t.Fatalf("err = %v, want ErrInsufficientConfirmations", err)
	}
	if invoice.paid != 0 {
		t.Fatal("invoice must not be marked paid on 0 confirmations")
	}

	if err := pay.UpdateConfirmations(1, now); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Handle(context.Background(), command.ConfirmPayment{
		MerchantID: "m-1",
		Network:    "evm:sepolia",
		TxHash:     "0xabc",
	}); err != nil {
		t.Fatal(err)
	}
	if invoice.paid != 1 {
		t.Fatalf("paid calls = %d, want 1", invoice.paid)
	}
}
