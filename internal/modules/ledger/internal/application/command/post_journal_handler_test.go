package command_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/eventbus"
)

type memAccounts struct {
	mu    sync.Mutex
	byKey map[string]*domain.Account
}

func (m *memAccounts) key(ot domain.OwnerType, oid string, k domain.AccountKind, ccy string) string {
	return string(ot) + "|" + oid + "|" + string(k) + "|" + ccy
}

func (m *memAccounts) FindByKey(_ context.Context, ot domain.OwnerType, oid string, k domain.AccountKind, ccy string) (*domain.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	acc, ok := m.byKey[m.key(ot, oid, k, ccy)]
	if !ok {
		return nil, domain.ErrAccountNotFound
	}
	return acc, nil
}

func (m *memAccounts) Save(_ context.Context, account *domain.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byKey == nil {
		m.byKey = map[string]*domain.Account{}
	}
	m.byKey[m.key(account.OwnerType(), account.OwnerID(), account.Kind(), account.Currency())] = account
	return nil
}

type memJournals struct {
	mu    sync.Mutex
	byKey map[string]*domain.Journal
}

func (m *memJournals) FindByIdempotencyKey(_ context.Context, key string) (*domain.Journal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.byKey[key]
	if !ok {
		return nil, domain.ErrJournalNotFound
	}
	return j, nil
}

func (m *memJournals) Save(_ context.Context, journal *domain.Journal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byKey == nil {
		m.byKey = map[string]*domain.Journal{}
	}
	m.byKey[journal.IdempotencyKey()] = journal
	return nil
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

func TestPostJournalIdempotent(t *testing.T) {
	accounts := &memAccounts{}
	journals := &memJournals{}
	pub := &memPub{}
	h := command.NewPostJournalHandler(accounts, journals, noopTX{}, pub)

	cmd := command.PostJournal{
		IdempotencyKey: "invoice_paid:inv-1",
		MerchantID:     "11111111-1111-1111-1111-111111111111",
		Amount:         "1.25",
		Currency:       "ETH",
		ReferenceType:  "invoice",
		ReferenceID:    "inv-1",
	}

	first, err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Created || first.JournalID == "" {
		t.Fatalf("first = %+v", first)
	}
	if len(pub.events) != 1 {
		t.Fatalf("events = %d, want 1", len(pub.events))
	}

	second, err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if second.Created {
		t.Fatal("second post should be idempotent noop")
	}
	if second.JournalID != first.JournalID {
		t.Fatalf("journal ids differ: %s vs %s", first.JournalID, second.JournalID)
	}
	if len(pub.events) != 1 {
		t.Fatalf("events = %d after retry, want 1", len(pub.events))
	}

	acc, err := accounts.FindByKey(
		context.Background(),
		domain.OwnerMerchant,
		cmd.MerchantID,
		domain.KindAvailable,
		"ETH",
	)
	if err != nil {
		t.Fatal(err)
	}
	if acc.Balance().Amount() != "1.25" {
		t.Fatalf("balance = %s, want 1.25", acc.Balance().Amount())
	}
}

func TestPostJournalRejectsBadAmount(t *testing.T) {
	h := command.NewPostJournalHandler(&memAccounts{}, &memJournals{}, noopTX{}, &memPub{})
	_, err := h.Handle(context.Background(), command.PostJournal{
		IdempotencyKey: "k",
		MerchantID:     "m",
		Amount:         "-1",
		Currency:       "ETH",
		ReferenceType:  "invoice",
		ReferenceID:    "i",
	})
	if !errors.Is(err, domain.ErrInvalidMoney) {
		t.Fatalf("err = %v", err)
	}
}

func TestPostDebitJournalCreditThenDebit(t *testing.T) {
	accounts := &memAccounts{}
	journals := &memJournals{}
	credit := command.NewPostJournalHandler(accounts, journals, noopTX{}, &memPub{})
	debit := command.NewPostDebitJournalHandler(accounts, journals, noopTX{}, &memPub{})
	merchantID := "11111111-1111-1111-1111-111111111111"

	if _, err := credit.Handle(context.Background(), command.PostJournal{
		IdempotencyKey: "invoice_paid:inv-1",
		MerchantID:     merchantID,
		Amount:         "2",
		Currency:       "ETH",
		ReferenceType:  "invoice",
		ReferenceID:    "inv-1",
	}); err != nil {
		t.Fatal(err)
	}

	res, err := debit.Handle(context.Background(), command.PostDebitJournal{
		IdempotencyKey: "withdrawal_debit:wd-1",
		MerchantID:     merchantID,
		Amount:         "1.5",
		Currency:       "ETH",
		ReferenceType:  "withdrawal",
		ReferenceID:    "wd-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created {
		t.Fatal("expected debit journal created")
	}

	acc, err := accounts.FindByKey(context.Background(), domain.OwnerMerchant, merchantID, domain.KindAvailable, "ETH")
	if err != nil {
		t.Fatal(err)
	}
	if acc.Balance().Amount() != "0.5" {
		t.Fatalf("balance = %s, want 0.5", acc.Balance().Amount())
	}

	again, err := debit.Handle(context.Background(), command.PostDebitJournal{
		IdempotencyKey: "withdrawal_debit:wd-1",
		MerchantID:     merchantID,
		Amount:         "1.5",
		Currency:       "ETH",
		ReferenceType:  "withdrawal",
		ReferenceID:    "wd-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if again.Created || again.JournalID != res.JournalID {
		t.Fatalf("idempotent retry failed: %+v", again)
	}
}

func TestPostDebitJournalInsufficientBalance(t *testing.T) {
	h := command.NewPostDebitJournalHandler(&memAccounts{}, &memJournals{}, noopTX{}, &memPub{})
	_, err := h.Handle(context.Background(), command.PostDebitJournal{
		IdempotencyKey: "withdrawal_debit:wd-2",
		MerchantID:     "11111111-1111-1111-1111-111111111111",
		Amount:         "1",
		Currency:       "ETH",
		ReferenceType:  "withdrawal",
		ReferenceID:    "wd-2",
	})
	if !errors.Is(err, domain.ErrInsufficientBalance) {
		t.Fatalf("err = %v, want ErrInsufficientBalance", err)
	}
}
