package command

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/payment"
)

type ConfirmTransactionHandler struct {
	repo      domain.TransactionRepository
	tx        ports.TransactionManager
	pub       ports.EventPublisher
	payment   ports.PaymentNotifier
	paymentMu sync.RWMutex
	now       func() time.Time
}

func NewConfirmTransactionHandler(
	repo domain.TransactionRepository,
	tx ports.TransactionManager,
	pub ports.EventPublisher,
) *ConfirmTransactionHandler {
	return &ConfirmTransactionHandler{
		repo: repo,
		tx:   tx,
		pub:  pub,
		now:  time.Now,
	}
}

func (h *ConfirmTransactionHandler) SetPaymentNotifier(n ports.PaymentNotifier) {
	h.paymentMu.Lock()
	defer h.paymentMu.Unlock()
	h.payment = n
}

type ConfirmTransactionResult struct {
	ID            string
	Network       string
	TxHash        string
	ToAddress     string
	Amount        string
	Currency      string
	Confirmations int
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (h *ConfirmTransactionHandler) Handle(ctx context.Context, cmd ConfirmTransaction) (ConfirmTransactionResult, error) {
	if err := domain.ValidateNetwork(cmd.Network); err != nil {
		return ConfirmTransactionResult{}, err
	}

	watched, err := h.repo.FindByNetworkTxHash(ctx, cmd.Network, cmd.TxHash)
	if err != nil {
		return ConfirmTransactionResult{}, err
	}

	var result ConfirmTransactionResult
	if watched.Status() == domain.TxStatusConfirmed {
		result = toConfirmResult(watched)
	} else {
		err = h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
			now := h.now().UTC()
			if err := watched.Confirm(now); err != nil {
				return err
			}
			if err := h.repo.Save(ctx, watched); err != nil {
				return fmt.Errorf("save confirmation: %w", err)
			}
			if err := h.pub.Publish(ctx, "blockchain", watched.ID(), transactionConfirmedEvent{
				Network:    watched.Network(),
				TxHash:     watched.TxHash(),
				OccurredAt: now,
			}); err != nil {
				return err
			}
			result = toConfirmResult(watched)
			return nil
		})
		if err != nil {
			return ConfirmTransactionResult{}, err
		}
	}

	notifier := h.paymentNotifier()
	if notifier == nil {
		return ConfirmTransactionResult{}, fmt.Errorf("payment notifier is not configured")
	}
	if _, err := notifier.Confirm(ctx, payment.ConfirmInput{
		MerchantID: cmd.MerchantID,
		Network:    cmd.Network,
		TxHash:     cmd.TxHash,
	}); err != nil {
		return ConfirmTransactionResult{}, fmt.Errorf("notify payment confirmed: %w", err)
	}
	return result, nil
}

func (h *ConfirmTransactionHandler) paymentNotifier() ports.PaymentNotifier {
	h.paymentMu.RLock()
	defer h.paymentMu.RUnlock()
	return h.payment
}

func toConfirmResult(t *domain.WatchedTransaction) ConfirmTransactionResult {
	return ConfirmTransactionResult{
		ID:            t.ID(),
		Network:       t.Network(),
		TxHash:        t.TxHash(),
		ToAddress:     t.ToAddress(),
		Amount:        t.Amount(),
		Currency:      t.Currency(),
		Confirmations: t.Confirmations(),
		Status:        t.Status().String(),
		CreatedAt:     t.CreatedAt(),
		UpdatedAt:     t.UpdatedAt(),
	}
}
