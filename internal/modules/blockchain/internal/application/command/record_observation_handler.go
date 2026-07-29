package command

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/payment"
)

type RecordObservationHandler struct {
	repo      domain.TransactionRepository
	tx        ports.TransactionManager
	pub       ports.EventPublisher
	payment   ports.PaymentNotifier
	paymentMu sync.RWMutex
	now       func() time.Time
}

func NewRecordObservationHandler(
	repo domain.TransactionRepository,
	tx ports.TransactionManager,
	pub ports.EventPublisher,
) *RecordObservationHandler {
	return &RecordObservationHandler{
		repo: repo,
		tx:   tx,
		pub:  pub,
		now:  time.Now,
	}
}

func (h *RecordObservationHandler) SetPaymentNotifier(n ports.PaymentNotifier) {
	h.paymentMu.Lock()
	defer h.paymentMu.Unlock()
	h.payment = n
}

type RecordObservationResult struct {
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

func (h *RecordObservationHandler) Handle(ctx context.Context, cmd RecordObservation) (RecordObservationResult, error) {
	if err := domain.ValidateNetwork(cmd.Network); err != nil {
		return RecordObservationResult{}, err
	}

	existing, err := h.repo.FindByNetworkTxHash(ctx, cmd.Network, cmd.TxHash)
	alreadyPersisted := err == nil
	if err != nil && !errors.Is(err, domain.ErrTxNotFound) {
		return RecordObservationResult{}, err
	}

	var result RecordObservationResult
	if alreadyPersisted {
		result = toObservationResult(existing)
	} else {
		err = h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
			existing, err := h.repo.FindByNetworkTxHash(ctx, cmd.Network, cmd.TxHash)
			if err == nil {
				result = toObservationResult(existing)
				return nil
			}
			if !errors.Is(err, domain.ErrTxNotFound) {
				return err
			}

			now := h.now().UTC()
			watched, err := domain.Observe(domain.ObserveParams{
				ID:            uuid.NewString(),
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
			if err := h.repo.Save(ctx, watched); err != nil {
				return fmt.Errorf("save observation: %w", err)
			}
			if err := h.pub.Publish(ctx, "blockchain", watched.ID(), transactionObservedEvent{
				Network:       watched.Network(),
				TxHash:        watched.TxHash(),
				ToAddress:     watched.ToAddress(),
				Amount:        watched.Amount(),
				Currency:      watched.Currency(),
				Confirmations: watched.Confirmations(),
				OccurredAt:    now,
			}); err != nil {
				return err
			}
			result = toObservationResult(watched)
			return nil
		})
		if err != nil {
			return RecordObservationResult{}, err
		}
	}

	notifier := h.paymentNotifier()
	if notifier == nil {
		return RecordObservationResult{}, fmt.Errorf("payment notifier is not configured")
	}
	if _, err := notifier.RecordObserved(ctx, payment.RecordObservedInput{
		MerchantID:    cmd.MerchantID,
		Network:       cmd.Network,
		TxHash:        cmd.TxHash,
		ToAddress:     cmd.ToAddress,
		Amount:        cmd.Amount,
		Currency:      cmd.Currency,
		Confirmations: cmd.Confirmations,
	}); err != nil {
		return RecordObservationResult{}, fmt.Errorf("notify payment observed: %w", err)
	}
	return result, nil
}

func (h *RecordObservationHandler) paymentNotifier() ports.PaymentNotifier {
	h.paymentMu.RLock()
	defer h.paymentMu.RUnlock()
	return h.payment
}

func toObservationResult(t *domain.WatchedTransaction) RecordObservationResult {
	return RecordObservationResult{
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
