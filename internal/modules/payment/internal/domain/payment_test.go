package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/domain"
)

func TestPaymentLifecycle(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)

	pay, err := domain.Detect(domain.DetectParams{
		ID:            domain.PaymentID("pay-1"),
		InvoiceID:     "inv-1",
		MerchantID:    "m-1",
		Network:       "evm:sepolia",
		TxHash:        "0xabc",
		ToAddress:     "stub:evm:sepolia:inv-1",
		Amount:        "0.01",
		Currency:      "ETH",
		Confirmations: 1,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if pay.Status() != domain.StatusDetected {
		t.Fatalf("status = %s, want detected", pay.Status())
	}

	t.Run("confirming then confirmed", func(t *testing.T) {
		p := mustClone(t, pay)
		if err := p.StartConfirming(now); err != nil {
			t.Fatal(err)
		}
		if p.Status() != domain.StatusConfirming {
			t.Fatalf("status = %s, want confirming", p.Status())
		}
		if err := p.Confirm(now); err != nil {
			t.Fatal(err)
		}
		if p.Status() != domain.StatusConfirmed {
			t.Fatalf("status = %s, want confirmed", p.Status())
		}
	})

	t.Run("confirm from detected", func(t *testing.T) {
		p := mustClone(t, pay)
		if err := p.Confirm(now); err != nil {
			t.Fatal(err)
		}
		if p.Status() != domain.StatusConfirmed {
			t.Fatalf("status = %s, want confirmed", p.Status())
		}
	})

	t.Run("invalid start confirming after confirm", func(t *testing.T) {
		p := mustClone(t, pay)
		if err := p.Confirm(now); err != nil {
			t.Fatal(err)
		}
		if err := p.StartConfirming(now); !errors.Is(err, domain.ErrInvalidTransition) {
			t.Fatalf("err = %v, want ErrInvalidTransition", err)
		}
	})

	t.Run("idempotent confirm when already confirmed", func(t *testing.T) {
		p := mustClone(t, pay)
		if err := p.Confirm(now); err != nil {
			t.Fatal(err)
		}
		if err := p.Confirm(now); err != nil {
			t.Fatalf("second confirm: %v", err)
		}
		if p.Status() != domain.StatusConfirmed {
			t.Fatalf("status = %s, want confirmed", p.Status())
		}
	})
}

func mustClone(t *testing.T, p *domain.Payment) *domain.Payment {
	t.Helper()
	return domain.Restore(
		p.ID(),
		p.InvoiceID(),
		p.MerchantID(),
		p.Network(),
		p.TxHash(),
		p.ToAddress(),
		p.Amount(),
		p.Currency(),
		p.Confirmations(),
		p.Status(),
		p.CreatedAt(),
		p.UpdatedAt(),
	)
}
