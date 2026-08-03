package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/domain"
)

func TestRequestAndComplete(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	w, err := domain.Request(domain.RequestParams{
		ID:             "wd-1",
		MerchantID:     "m-1",
		IdempotencyKey: "key-1",
		Amount:         "1.5",
		Currency:       "ETH",
		Network:        "evm:sepolia",
		ToAddress:      "0xabc",
		JournalID:      "j-1",
		Now:            now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if w.Status() != domain.StatusRequested {
		t.Fatalf("status = %s", w.Status())
	}

	if err := w.Complete("0xtx", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if w.Status() != domain.StatusCompleted || w.TxHash() != "0xtx" {
		t.Fatalf("status=%s tx=%s", w.Status(), w.TxHash())
	}

	if err := w.Complete("0xother", now); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("err = %v, want ErrInvalidTransition", err)
	}
}

func TestRequestRejectsInvalidAmount(t *testing.T) {
	now := time.Now().UTC()
	_, err := domain.Request(domain.RequestParams{
		ID:             "wd-1",
		MerchantID:     "m-1",
		IdempotencyKey: "key-1",
		Amount:         "0",
		Currency:       "ETH",
		Network:        "evm:sepolia",
		ToAddress:      "0xabc",
		JournalID:      "j-1",
		Now:            now,
	})
	if !errors.Is(err, domain.ErrInvalidWithdrawal) {
		t.Fatalf("err = %v", err)
	}
}

func TestRejectFromRequested(t *testing.T) {
	now := time.Now().UTC()
	w, err := domain.Request(domain.RequestParams{
		ID:             "wd-1",
		MerchantID:     "m-1",
		IdempotencyKey: "key-1",
		Amount:         "1",
		Currency:       "ETH",
		Network:        "evm:sepolia",
		ToAddress:      "0xabc",
		JournalID:      "j-1",
		Now:            now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Reject(now); err != nil {
		t.Fatal(err)
	}
	if w.Status() != domain.StatusRejected {
		t.Fatalf("status = %s", w.Status())
	}
}
