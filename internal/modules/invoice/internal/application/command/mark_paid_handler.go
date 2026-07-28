package command

import (
	"context"
	"fmt"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/domain"
)

type MarkPaidHandler struct {
	repo domain.Repository
	tx   ports.TransactionManager
	pub  ports.EventPublisher
	now  func() time.Time
}

func NewMarkPaidHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	pub ports.EventPublisher,
) *MarkPaidHandler {
	return &MarkPaidHandler{
		repo: repo,
		tx:   tx,
		pub:  pub,
		now:  time.Now,
	}
}

func (h *MarkPaidHandler) Handle(ctx context.Context, cmd MarkPaid) error {
	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		inv, err := h.repo.FindByID(ctx, domain.InvoiceID(cmd.InvoiceID))
		if err != nil {
			return err
		}

		now := h.now().UTC()
		if err := inv.MarkPaid(cmd.TxHash, now); err != nil {
			return err
		}
		if err := h.repo.Save(ctx, inv); err != nil {
			return fmt.Errorf("save invoice: %w", err)
		}

		return h.pub.Publish(ctx, "invoice", inv.ID().String(), invoicePaidEvent{
			InvoiceID:  inv.ID().String(),
			MerchantID: inv.MerchantID(),
			TxHash:     inv.TxHash(),
			OccurredAt: now,
		})
	})
}
