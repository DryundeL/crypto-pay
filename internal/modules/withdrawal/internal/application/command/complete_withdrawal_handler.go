package command

import (
	"context"
	"fmt"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/domain"
)

type CompleteWithdrawalHandler struct {
	repo domain.Repository
	tx   ports.TransactionManager
	pub  ports.EventPublisher
	now  func() time.Time
}

func NewCompleteWithdrawalHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	pub ports.EventPublisher,
) *CompleteWithdrawalHandler {
	return &CompleteWithdrawalHandler{
		repo: repo,
		tx:   tx,
		pub:  pub,
		now:  time.Now,
	}
}

func (h *CompleteWithdrawalHandler) Handle(ctx context.Context, cmd CompleteWithdrawal) (CompleteWithdrawalResult, error) {
	var result CompleteWithdrawalResult
	err := h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		w, err := h.repo.FindByID(ctx, domain.WithdrawalID(cmd.WithdrawalID))
		if err != nil {
			return err
		}

		if w.Status() == domain.StatusCompleted {
			if w.TxHash() == cmd.TxHash || cmd.TxHash == "" {
				result = toCompleteResult(w)
				return nil
			}
			return fmt.Errorf("%w: already completed with different tx_hash", domain.ErrInvalidTransition)
		}

		now := h.now().UTC()
		if err := w.Complete(cmd.TxHash, now); err != nil {
			return err
		}
		if err := h.repo.Save(ctx, w); err != nil {
			return fmt.Errorf("save withdrawal: %w", err)
		}
		if err := h.pub.Publish(ctx, "withdrawal", w.ID().String(), withdrawalCompletedEvent{
			WithdrawalID: w.ID().String(),
			MerchantID:   w.MerchantID(),
			TxHash:       w.TxHash(),
			OccurredAt:   now,
		}); err != nil {
			return err
		}
		result = toCompleteResult(w)
		return nil
	})
	return result, err
}

func toCompleteResult(w *domain.Withdrawal) CompleteWithdrawalResult {
	return CompleteWithdrawalResult{
		ID:         w.ID().String(),
		MerchantID: w.MerchantID(),
		Status:     w.Status().String(),
		TxHash:     w.TxHash(),
		Amount:     w.Amount(),
		Currency:   w.Currency(),
		Network:    w.Network(),
		ToAddress:  w.ToAddress(),
		JournalID:  w.JournalID(),
		UpdatedAt:  w.UpdatedAt(),
	}
}
