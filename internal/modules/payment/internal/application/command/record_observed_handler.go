package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/domain"
)

type RecordObservedHandler struct {
	repo    domain.Repository
	tx      ports.TransactionManager
	pub     ports.EventPublisher
	lookup  ports.InvoiceLookup
	invoice ports.InvoiceLifecycle
	now     func() time.Time
}

func NewRecordObservedHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	pub ports.EventPublisher,
	lookup ports.InvoiceLookup,
	invoice ports.InvoiceLifecycle,
) *RecordObservedHandler {
	return &RecordObservedHandler{
		repo:    repo,
		tx:      tx,
		pub:     pub,
		lookup:  lookup,
		invoice: invoice,
		now:     time.Now,
	}
}

type RecordObservedResult struct {
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

func (h *RecordObservedHandler) Handle(ctx context.Context, cmd RecordObserved) (RecordObservedResult, error) {
	if err := validateNetwork(cmd.Network); err != nil {
		return RecordObservedResult{}, err
	}

	existing, err := h.repo.FindByNetworkTxHash(ctx, cmd.Network, cmd.TxHash)
	if err == nil {
		if existing.MerchantID() != cmd.MerchantID {
			return RecordObservedResult{}, domain.ErrPaymentNotFound
		}
		return h.refreshConfirmations(ctx, existing, cmd.Confirmations)
	}
	if !errors.Is(err, domain.ErrPaymentNotFound) {
		return RecordObservedResult{}, err
	}

	ref, err := h.lookup.FindPendingByAddress(ctx, cmd.Network, cmd.ToAddress)
	if err != nil {
		if errors.Is(err, invoice.ErrNotFound) {
			return RecordObservedResult{}, domain.ErrPaymentNotFound
		}
		return RecordObservedResult{}, err
	}
	if ref.MerchantID != cmd.MerchantID {
		return RecordObservedResult{}, domain.ErrPaymentNotFound
	}
	if ref.Amount != cmd.Amount || ref.Currency != cmd.Currency {
		return RecordObservedResult{}, fmt.Errorf("%w: amount/currency mismatch", domain.ErrInvalidPayment)
	}

	var result RecordObservedResult
	err = h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		now := h.now().UTC()
		pay, err := domain.Detect(domain.DetectParams{
			ID:            domain.PaymentID(uuid.NewString()),
			InvoiceID:     ref.ID,
			MerchantID:    ref.MerchantID,
			Network:       cmd.Network,
			TxHash:        cmd.TxHash,
			ToAddress:     cmd.ToAddress,
			Amount:        cmd.Amount,
			Currency:      cmd.Currency,
			Confirmations: cmd.Confirmations,
			Now:           now,
		})
		if err != nil {
			return err
		}
		if err := pay.StartConfirming(now); err != nil {
			return err
		}
		if err := h.repo.Save(ctx, pay); err != nil {
			return fmt.Errorf("save payment: %w", err)
		}
		if err := h.pub.Publish(ctx, "payment", pay.ID().String(), paymentDetectedEvent{
			PaymentID:  pay.ID().String(),
			InvoiceID:  pay.InvoiceID(),
			TxHash:     pay.TxHash(),
			Network:    pay.Network(),
			OccurredAt: now,
		}); err != nil {
			return err
		}
		result = toObservedResult(pay)
		return nil
	})
	if err != nil {
		return RecordObservedResult{}, err
	}

	if err := h.invoice.MarkConfirming(ctx, result.InvoiceID); err != nil {
		return RecordObservedResult{}, fmt.Errorf("mark invoice confirming: %w", err)
	}
	return result, nil
}

func (h *RecordObservedHandler) refreshConfirmations(
	ctx context.Context,
	pay *domain.Payment,
	confirmations int,
) (RecordObservedResult, error) {
	if confirmations <= pay.Confirmations() {
		return toObservedResult(pay), nil
	}

	err := h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		now := h.now().UTC()
		if err := pay.UpdateConfirmations(confirmations, now); err != nil {
			return err
		}
		if err := h.repo.Save(ctx, pay); err != nil {
			return fmt.Errorf("save payment confirmations: %w", err)
		}
		return nil
	})
	if err != nil {
		return RecordObservedResult{}, err
	}
	return toObservedResult(pay), nil
}

func toObservedResult(p *domain.Payment) RecordObservedResult {
	return RecordObservedResult{
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

func validateNetwork(network string) error {
	// Keep in sync with blockchain.Network* public constants (avoid import cycle).
	switch network {
	case "evm:sepolia", "btc:regtest", "btc:testnet":
		return nil
	default:
		return fmt.Errorf("%w: unsupported network", domain.ErrInvalidPayment)
	}
}
