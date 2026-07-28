package command

import (
	"context"
	"fmt"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/domain"
)

type MarkConfirmingHandler struct {
	repo domain.Repository
	tx   ports.TransactionManager
	now  func() time.Time
}

func NewMarkConfirmingHandler(repo domain.Repository, tx ports.TransactionManager) *MarkConfirmingHandler {
	return &MarkConfirmingHandler{
		repo: repo,
		tx:   tx,
		now:  time.Now,
	}
}

func (h *MarkConfirmingHandler) Handle(ctx context.Context, cmd MarkConfirming) error {
	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		inv, err := h.repo.FindByID(ctx, domain.InvoiceID(cmd.InvoiceID))
		if err != nil {
			return err
		}
		if err := inv.MarkConfirming(h.now()); err != nil {
			return err
		}
		if err := h.repo.Save(ctx, inv); err != nil {
			return fmt.Errorf("save invoice: %w", err)
		}
		return nil
	})
}
