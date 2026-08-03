package command_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/eventbus"
)

type memRepo struct {
	mu    sync.Mutex
	byID  map[string]*domain.Withdrawal
	byKey map[string]*domain.Withdrawal
}

func (r *memRepo) Save(_ context.Context, w *domain.Withdrawal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.byID == nil {
		r.byID = map[string]*domain.Withdrawal{}
		r.byKey = map[string]*domain.Withdrawal{}
	}
	r.byID[w.ID().String()] = w
	r.byKey[w.IdempotencyKey()] = w
	return nil
}

func (r *memRepo) FindByID(_ context.Context, id domain.WithdrawalID) (*domain.Withdrawal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.byID[id.String()]
	if !ok {
		return nil, domain.ErrWithdrawalNotFound
	}
	return w, nil
}

func (r *memRepo) FindByIdempotencyKey(_ context.Context, key string) (*domain.Withdrawal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.byKey[key]
	if !ok {
		return nil, domain.ErrWithdrawalNotFound
	}
	return w, nil
}

type noopTX struct{}

func (noopTX) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type memPub struct {
	n int
}

func (p *memPub) Publish(_ context.Context, _, _ string, _ eventbus.IntegrationEvent) error {
	p.n++
	return nil
}

type memLedger struct {
	calls int
	fail  error
	bal   string
}

func (l *memLedger) PostDebitJournal(_ context.Context, in ledger.PostDebitJournalInput) (ledger.PostDebitJournalResult, error) {
	l.calls++
	if l.fail != nil {
		return ledger.PostDebitJournalResult{}, l.fail
	}
	if l.bal == "0" {
		return ledger.PostDebitJournalResult{}, ledger.ErrInsufficientBalance
	}
	return ledger.PostDebitJournalResult{JournalID: "j-" + in.ReferenceID, Created: l.calls == 1}, nil
}

func TestRequestWithdrawalDebitsAndSaves(t *testing.T) {
	repo := &memRepo{}
	pub := &memPub{}
	ld := &memLedger{bal: "10"}
	h := command.NewRequestWithdrawalHandler(repo, noopTX{}, pub, ld)

	res, err := h.Handle(context.Background(), command.RequestWithdrawal{
		MerchantID:     "m-1",
		IdempotencyKey: "key-1",
		Amount:         "1.5",
		Currency:       "ETH",
		Network:        "evm:sepolia",
		ToAddress:      "0xabc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created || res.JournalID == "" || res.Status != "requested" {
		t.Fatalf("result = %+v", res)
	}
	if ld.calls != 1 || pub.n != 1 {
		t.Fatalf("debit calls=%d pub=%d", ld.calls, pub.n)
	}

	again, err := h.Handle(context.Background(), command.RequestWithdrawal{
		MerchantID:     "m-1",
		IdempotencyKey: "key-1",
		Amount:         "1.5",
		Currency:       "ETH",
		Network:        "evm:sepolia",
		ToAddress:      "0xabc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if again.Created || again.ID != res.ID {
		t.Fatalf("idempotent = %+v", again)
	}
	if ld.calls != 1 {
		t.Fatalf("debit should not be called again, calls=%d", ld.calls)
	}
}

func TestRequestWithdrawalInsufficientBalance(t *testing.T) {
	repo := &memRepo{}
	h := command.NewRequestWithdrawalHandler(repo, noopTX{}, &memPub{}, &memLedger{bal: "0"})

	_, err := h.Handle(context.Background(), command.RequestWithdrawal{
		MerchantID:     "m-1",
		IdempotencyKey: "key-2",
		Amount:         "1",
		Currency:       "ETH",
		Network:        "evm:sepolia",
		ToAddress:      "0xabc",
	})
	if !errors.Is(err, ledger.ErrInsufficientBalance) {
		t.Fatalf("err = %v", err)
	}
	if _, err := repo.FindByIdempotencyKey(context.Background(), "key-2"); !errors.Is(err, domain.ErrWithdrawalNotFound) {
		t.Fatal("withdrawal should not be persisted")
	}
}

func TestCompleteWithdrawal(t *testing.T) {
	repo := &memRepo{}
	pub := &memPub{}
	req := command.NewRequestWithdrawalHandler(repo, noopTX{}, pub, &memLedger{bal: "10"})
	created, err := req.Handle(context.Background(), command.RequestWithdrawal{
		MerchantID:     "m-1",
		IdempotencyKey: "key-3",
		Amount:         "1",
		Currency:       "ETH",
		Network:        "evm:sepolia",
		ToAddress:      "0xabc",
	})
	if err != nil {
		t.Fatal(err)
	}

	complete := command.NewCompleteWithdrawalHandler(repo, noopTX{}, pub)
	res, err := complete.Handle(context.Background(), command.CompleteWithdrawal{
		WithdrawalID: created.ID,
		TxHash:       "0xdead",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "completed" || res.TxHash != "0xdead" {
		t.Fatalf("result = %+v", res)
	}
}
