package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/domain"
)

type RequestWithdrawalHandler struct {
	repo   domain.Repository
	tx     ports.TransactionManager
	pub    ports.EventPublisher
	ledger ports.LedgerDebiter
	now    func() time.Time
	newID  func() string
}

func NewRequestWithdrawalHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	pub ports.EventPublisher,
	ledger ports.LedgerDebiter,
) *RequestWithdrawalHandler {
	return &RequestWithdrawalHandler{
		repo:   repo,
		tx:     tx,
		pub:    pub,
		ledger: ledger,
		now:    time.Now,
		newID:  func() string { return uuid.NewString() },
	}
}

func (h *RequestWithdrawalHandler) Handle(ctx context.Context, cmd RequestWithdrawal) (RequestWithdrawalResult, error) {
	existing, err := h.repo.FindByIdempotencyKey(ctx, cmd.IdempotencyKey)
	if err == nil {
		return toRequestResult(existing, false), nil
	}
	if !errors.Is(err, domain.ErrWithdrawalNotFound) {
		return RequestWithdrawalResult{}, err
	}

	now := h.now().UTC()
	id := h.newID()

	// Validate before debit so we don't take funds on bad input.
	if _, err := domain.Request(domain.RequestParams{
		ID:             domain.WithdrawalID(id),
		MerchantID:     cmd.MerchantID,
		IdempotencyKey: cmd.IdempotencyKey,
		Amount:         cmd.Amount,
		Currency:       cmd.Currency,
		Network:        cmd.Network,
		ToAddress:      cmd.ToAddress,
		JournalID:      "pending",
		Now:            now,
	}); err != nil {
		return RequestWithdrawalResult{}, err
	}

	debit, err := h.ledger.PostDebitJournal(ctx, ledger.PostDebitJournalInput{
		IdempotencyKey: "withdrawal_debit:" + cmd.IdempotencyKey,
		MerchantID:     cmd.MerchantID,
		Amount:         cmd.Amount,
		Currency:       cmd.Currency,
		ReferenceType:  "withdrawal",
		ReferenceID:    id,
	})
	if err != nil {
		return RequestWithdrawalResult{}, err
	}

	w, err := domain.Request(domain.RequestParams{
		ID:             domain.WithdrawalID(id),
		MerchantID:     cmd.MerchantID,
		IdempotencyKey: cmd.IdempotencyKey,
		Amount:         cmd.Amount,
		Currency:       cmd.Currency,
		Network:        cmd.Network,
		ToAddress:      cmd.ToAddress,
		JournalID:      debit.JournalID,
		Now:            now,
	})
	if err != nil {
		return RequestWithdrawalResult{}, err
	}

	created := true
	err = h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if again, err := h.repo.FindByIdempotencyKey(ctx, cmd.IdempotencyKey); err == nil {
			w = again
			created = false
			return nil
		} else if !errors.Is(err, domain.ErrWithdrawalNotFound) {
			return err
		}

		if err := h.repo.Save(ctx, w); err != nil {
			return fmt.Errorf("save withdrawal: %w", err)
		}
		return h.pub.Publish(ctx, "withdrawal", w.ID().String(), withdrawalRequestedEvent{
			WithdrawalID: w.ID().String(),
			MerchantID:   w.MerchantID(),
			Amount:       w.Amount(),
			Currency:     w.Currency(),
			Network:      w.Network(),
			OccurredAt:   now,
		})
	})
	if err != nil {
		return RequestWithdrawalResult{}, err
	}
	return toRequestResult(w, created), nil
}

func toRequestResult(w *domain.Withdrawal, created bool) RequestWithdrawalResult {
	return RequestWithdrawalResult{
		ID:             w.ID().String(),
		MerchantID:     w.MerchantID(),
		IdempotencyKey: w.IdempotencyKey(),
		Amount:         w.Amount(),
		Currency:       w.Currency(),
		Network:        w.Network(),
		ToAddress:      w.ToAddress(),
		Status:         w.Status().String(),
		JournalID:      w.JournalID(),
		Created:        created,
		CreatedAt:      w.CreatedAt(),
		UpdatedAt:      w.UpdatedAt(),
	}
}
