package command

import (
	"context"
	"fmt"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/domain"
)

type CancelInvoiceHandler struct {
	repo domain.Repository
	tx   ports.TransactionManager
	now  func() time.Time
}

func NewCancelInvoiceHandler(repo domain.Repository, tx ports.TransactionManager) *CancelInvoiceHandler {
	return &CancelInvoiceHandler{
		repo: repo,
		tx:   tx,
		now:  time.Now,
	}
}

type CancelInvoiceResult struct {
	ID         string
	MerchantID string
	Status     string
	Amount     string
	Currency   string
	Network    string
	Address    string
	TxHash     string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (h *CancelInvoiceHandler) Handle(ctx context.Context, cmd CancelInvoice) (CancelInvoiceResult, error) {
	var result CancelInvoiceResult

	err := h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		inv, err := h.repo.FindByID(ctx, domain.InvoiceID(cmd.InvoiceID))
		if err != nil {
			return err
		}
		// Hide existence of invoices belonging to other merchants.
		if inv.MerchantID() != cmd.MerchantID {
			return domain.ErrInvoiceNotFound
		}

		if err := inv.Cancel(h.now()); err != nil {
			return err
		}
		if err := h.repo.Save(ctx, inv); err != nil {
			return fmt.Errorf("save invoice: %w", err)
		}

		result = toCancelResult(inv)
		return nil
	})
	if err != nil {
		return CancelInvoiceResult{}, err
	}
	return result, nil
}

func toCancelResult(inv *domain.Invoice) CancelInvoiceResult {
	return CancelInvoiceResult{
		ID:         inv.ID().String(),
		MerchantID: inv.MerchantID(),
		Status:     inv.Status().String(),
		Amount:     inv.Amount().Amount(),
		Currency:   inv.Amount().Currency(),
		Network:    inv.Network(),
		Address:    inv.Address(),
		TxHash:     inv.TxHash(),
		ExpiresAt:  inv.ExpiresAt(),
		CreatedAt:  inv.CreatedAt(),
		UpdatedAt:  inv.UpdatedAt(),
	}
}
