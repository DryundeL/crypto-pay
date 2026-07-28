package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/domain"
)

func TestInvoiceLifecycle(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	money, err := domain.NewMoney("1.5", "ETH")
	if err != nil {
		t.Fatal(err)
	}

	inv, err := domain.Create(domain.CreateParams{
		ID:         domain.InvoiceID("inv-1"),
		MerchantID: "m-1",
		Amount:     money,
		Network:    "evm:sepolia",
		Address:    "stub:evm:sepolia:inv-1",
		ExpiresAt:  now.Add(30 * time.Minute),
		Now:        now,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if inv.Status() != domain.StatusPending {
		t.Fatalf("status = %s, want pending", inv.Status())
	}

	t.Run("cancel from pending", func(t *testing.T) {
		c := mustClone(t, inv)
		if err := c.Cancel(now); err != nil {
			t.Fatal(err)
		}
		if c.Status() != domain.StatusCancelled {
			t.Fatalf("status = %s, want cancelled", c.Status())
		}
	})

	t.Run("expire before TTL fails", func(t *testing.T) {
		c := mustClone(t, inv)
		if err := c.Expire(now.Add(10 * time.Minute)); !errors.Is(err, domain.ErrInvoiceNotExpired) {
			t.Fatalf("err = %v, want ErrInvoiceNotExpired", err)
		}
	})

	t.Run("expire after TTL", func(t *testing.T) {
		c := mustClone(t, inv)
		if err := c.Expire(now.Add(31 * time.Minute)); err != nil {
			t.Fatal(err)
		}
		if c.Status() != domain.StatusExpired {
			t.Fatalf("status = %s, want expired", c.Status())
		}
	})

	t.Run("confirming then paid", func(t *testing.T) {
		c := mustClone(t, inv)
		if err := c.MarkConfirming(now); err != nil {
			t.Fatal(err)
		}
		if c.Status() != domain.StatusConfirming {
			t.Fatalf("status = %s, want confirming", c.Status())
		}
		if err := c.MarkPaid("0xabc", now); err != nil {
			t.Fatal(err)
		}
		if c.Status() != domain.StatusPaid {
			t.Fatalf("status = %s, want paid", c.Status())
		}
		if c.TxHash() != "0xabc" {
			t.Fatalf("tx_hash = %s", c.TxHash())
		}
	})

	t.Run("paid from pending", func(t *testing.T) {
		c := mustClone(t, inv)
		if err := c.MarkPaid("0xdef", now); err != nil {
			t.Fatal(err)
		}
		if c.Status() != domain.StatusPaid {
			t.Fatalf("status = %s, want paid", c.Status())
		}
	})

	t.Run("invalid transitions", func(t *testing.T) {
		c := mustClone(t, inv)
		if err := c.Cancel(now); err != nil {
			t.Fatal(err)
		}
		if err := c.MarkConfirming(now); !errors.Is(err, domain.ErrInvalidTransition) {
			t.Fatalf("err = %v, want ErrInvalidTransition", err)
		}
		if err := c.MarkPaid("0x", now); !errors.Is(err, domain.ErrInvalidTransition) {
			t.Fatalf("err = %v, want ErrInvalidTransition", err)
		}
		if err := c.Expire(now.Add(time.Hour)); !errors.Is(err, domain.ErrInvalidTransition) {
			t.Fatalf("err = %v, want ErrInvalidTransition", err)
		}
	})
}

func mustClone(t *testing.T, inv *domain.Invoice) *domain.Invoice {
	t.Helper()
	return domain.Restore(
		inv.ID(),
		inv.MerchantID(),
		inv.Status(),
		inv.Amount(),
		inv.Network(),
		inv.Address(),
		inv.TxHash(),
		inv.ExpiresAt(),
		inv.CreatedAt(),
		inv.UpdatedAt(),
	)
}
