package command

import (
	"context"
	"fmt"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/domain"
)

type ConfirmPaymentHandler struct {
	repo    domain.Repository
	tx      ports.TransactionManager
	pub     ports.EventPublisher
	invoice ports.InvoiceLifecycle
	now     func() time.Time
}

func NewConfirmPaymentHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	pub ports.EventPublisher,
	invoice ports.InvoiceLifecycle,
) *ConfirmPaymentHandler {
	return &ConfirmPaymentHandler{
		repo:    repo,
		tx:      tx,
		pub:     pub,
		invoice: invoice,
		now:     time.Now,
	}
}

type ConfirmPaymentResult struct {
	ID            string
	InvoiceID     string
	MerchantID    string
	Status        string
	Network       string
	TxHash        string
	ToAddress     string
	Amount        string
	Currency      string
	Confirmations int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (h *ConfirmPaymentHandler) Handle(ctx context.Context, cmd ConfirmPayment) (ConfirmPaymentResult, error) {
	if err := validateNetwork(cmd.Network); err != nil {
		return ConfirmPaymentResult{}, err
	}

	pay, err := h.repo.FindByNetworkTxHash(ctx, cmd.Network, cmd.TxHash)
	if err != nil {
		return ConfirmPaymentResult{}, err
	}
	if pay.MerchantID() != cmd.MerchantID {
		return ConfirmPaymentResult{}, domain.ErrPaymentNotFound
	}

	if pay.Status() == domain.StatusConfirmed {
		return toConfirmResult(pay), nil
	}

	var result ConfirmPaymentResult
	err = h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		now := h.now().UTC()
		if err := pay.Confirm(now); err != nil {
			return err
		}
		if err := h.repo.Save(ctx, pay); err != nil {
			return fmt.Errorf("save payment: %w", err)
		}
		if err := h.pub.Publish(ctx, "payment", pay.ID().String(), paymentConfirmedEvent{
			PaymentID:  pay.ID().String(),
			InvoiceID:  pay.InvoiceID(),
			TxHash:     pay.TxHash(),
			OccurredAt: now,
		}); err != nil {
			return err
		}
		result = toConfirmResult(pay)
		return nil
	})
	if err != nil {
		return ConfirmPaymentResult{}, err
	}

	if err := h.invoice.MarkPaid(ctx, result.InvoiceID, result.TxHash); err != nil {
		return ConfirmPaymentResult{}, fmt.Errorf("mark invoice paid: %w", err)
	}
	return result, nil
}

func toConfirmResult(p *domain.Payment) ConfirmPaymentResult {
	return ConfirmPaymentResult{
		ID:            p.ID().String(),
		InvoiceID:     p.InvoiceID(),
		MerchantID:    p.MerchantID(),
		Status:        p.Status().String(),
		Network:       p.Network(),
		TxHash:        p.TxHash(),
		ToAddress:     p.ToAddress(),
		Amount:        p.Amount(),
		Currency:      p.Currency(),
		Confirmations: p.Confirmations(),
		CreatedAt:     p.CreatedAt(),
		UpdatedAt:     p.UpdatedAt(),
	}
}
