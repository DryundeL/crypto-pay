package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/domain"
)

func TestMoneyAddSub(t *testing.T) {
	a, err := domain.NewMoney("1.5", "ETH")
	if err != nil {
		t.Fatal(err)
	}
	b, err := domain.NewMoney("0.25", "ETH")
	if err != nil {
		t.Fatal(err)
	}

	sum, err := a.Add(b)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Amount() != "1.75" {
		t.Fatalf("sum = %s, want 1.75", sum.Amount())
	}

	diff, err := a.Sub(b)
	if err != nil {
		t.Fatal(err)
	}
	if diff.Amount() != "1.25" {
		t.Fatalf("diff = %s, want 1.25", diff.Amount())
	}
}

func TestMoneyRejectsInvalid(t *testing.T) {
	cases := []string{"", "-1", "1e3", "abc", "01.2.3"}
	for _, amount := range cases {
		if _, err := domain.NewMoney(amount, "ETH"); !errors.Is(err, domain.ErrInvalidMoney) {
			t.Fatalf("amount %q: err = %v, want ErrInvalidMoney", amount, err)
		}
	}
}

func TestPostJournalBalanced(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	money, err := domain.NewMoney("10", "USDT")
	if err != nil {
		t.Fatal(err)
	}

	j, err := domain.Post(domain.PostParams{
		ID:                "j-1",
		IdempotencyKey:    "invoice_paid:inv-1",
		MerchantID:        "m-1",
		Amount:            money,
		ReferenceType:     "invoice",
		ReferenceID:       "inv-1",
		ClearingAccountID: "acc-clearing",
		MerchantAccountID: "acc-merchant",
		ClearingEntryID:   "e-1",
		MerchantEntryID:   "e-2",
		Now:               now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if j.ID().String() != "j-1" {
		t.Fatalf("id = %s", j.ID())
	}
	entries := j.Entries()
	if len(entries) != 2 {
		t.Fatalf("entries = %d", len(entries))
	}
	if entries[0].Side() != domain.SideDebit || entries[1].Side() != domain.SideCredit {
		t.Fatalf("sides = %s/%s", entries[0].Side(), entries[1].Side())
	}
	if j.Amount().Amount() != "10" || j.Currency() != "USDT" {
		t.Fatalf("amount = %s %s", j.Amount().Amount(), j.Currency())
	}
}

func TestPostJournalRequiresFields(t *testing.T) {
	now := time.Now().UTC()
	money, _ := domain.NewMoney("1", "ETH")

	_, err := domain.Post(domain.PostParams{
		ID:                "j-1",
		IdempotencyKey:    "",
		MerchantID:        "m-1",
		Amount:            money,
		ReferenceType:     "invoice",
		ReferenceID:       "inv-1",
		ClearingAccountID: "a1",
		MerchantAccountID: "a2",
		ClearingEntryID:   "e1",
		MerchantEntryID:   "e2",
		Now:               now,
	})
	if !errors.Is(err, domain.ErrInvalidJournal) {
		t.Fatalf("err = %v, want ErrInvalidJournal", err)
	}
}

func TestAccountApplyCreditIncreasesAvailable(t *testing.T) {
	now := time.Now().UTC()
	acc, err := domain.CreateAccount(domain.CreateAccountParams{
		ID:        "acc-1",
		OwnerType: domain.OwnerMerchant,
		OwnerID:   "m-1",
		Kind:      domain.KindAvailable,
		Currency:  "ETH",
		Now:       now,
	})
	if err != nil {
		t.Fatal(err)
	}

	amount, err := domain.NewMoney("2.5", "ETH")
	if err != nil {
		t.Fatal(err)
	}
	if err := acc.ApplyEntry(domain.SideCredit, amount, now); err != nil {
		t.Fatal(err)
	}
	if acc.Balance().Amount() != "2.5" {
		t.Fatalf("balance = %s, want 2.5", acc.Balance().Amount())
	}

	if err := acc.ApplyEntry(domain.SideDebit, amount, now); err != nil {
		t.Fatal(err)
	}
	if !acc.Balance().IsZero() {
		t.Fatalf("balance = %s, want 0", acc.Balance().Amount())
	}

	if err := acc.ApplyEntry(domain.SideDebit, amount, now); !errors.Is(err, domain.ErrInsufficientBalance) {
		t.Fatalf("err = %v, want ErrInsufficientBalance", err)
	}
}

func TestPostDebitJournalBalanced(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	money, err := domain.NewMoney("3", "USDT")
	if err != nil {
		t.Fatal(err)
	}

	j, err := domain.PostDebit(domain.PostParams{
		ID:                "j-debit",
		IdempotencyKey:    "withdrawal_debit:wd-1",
		MerchantID:        "m-1",
		Amount:            money,
		ReferenceType:     "withdrawal",
		ReferenceID:       "wd-1",
		ClearingAccountID: "acc-clearing",
		MerchantAccountID: "acc-merchant",
		ClearingEntryID:   "e-1",
		MerchantEntryID:   "e-2",
		Now:               now,
	})
	if err != nil {
		t.Fatal(err)
	}
	entries := j.Entries()
	if entries[0].Side() != domain.SideDebit || entries[0].AccountID() != "acc-merchant" {
		t.Fatalf("first entry = %s %s", entries[0].Side(), entries[0].AccountID())
	}
	if entries[1].Side() != domain.SideCredit || entries[1].AccountID() != "acc-clearing" {
		t.Fatalf("second entry = %s %s", entries[1].Side(), entries[1].AccountID())
	}
}

func TestClearingDebitIncreasesBalance(t *testing.T) {
	now := time.Now().UTC()
	acc, err := domain.CreateAccount(domain.CreateAccountParams{
		ID:        "acc-c",
		OwnerType: domain.OwnerSystem,
		OwnerID:   domain.SystemOwnerID,
		Kind:      domain.KindClearing,
		Currency:  "ETH",
		Now:       now,
	})
	if err != nil {
		t.Fatal(err)
	}
	amount, _ := domain.NewMoney("1", "ETH")
	if err := acc.ApplyEntry(domain.SideDebit, amount, now); err != nil {
		t.Fatal(err)
	}
	if acc.Balance().Amount() != "1" {
		t.Fatalf("balance = %s, want 1", acc.Balance().Amount())
	}
}
