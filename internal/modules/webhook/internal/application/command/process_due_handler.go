package command

import (
	"context"
	"fmt"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/domain"
)

type ProcessDueHandler struct {
	repo      domain.Repository
	tx        ports.TransactionManager
	pub       ports.EventPublisher
	deliverer ports.HTTPDeliverer
	now       func() time.Time
}

func NewProcessDueHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	pub ports.EventPublisher,
	deliverer ports.HTTPDeliverer,
) *ProcessDueHandler {
	return &ProcessDueHandler{
		repo:      repo,
		tx:        tx,
		pub:       pub,
		deliverer: deliverer,
		now:       time.Now,
	}
}

func (h *ProcessDueHandler) Handle(ctx context.Context, cmd ProcessDue) (ProcessDueResult, error) {
	limit := cmd.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	now := h.now().UTC()
	due, err := h.repo.FindDue(ctx, now, limit)
	if err != nil {
		return ProcessDueResult{}, err
	}

	var result ProcessDueResult
	for _, d := range due {
		outcome, err := h.attemptOne(ctx, d)
		if err != nil {
			return result, err
		}
		result.Processed++
		switch outcome {
		case "sent":
			result.Sent++
		case "failed":
			result.Failed++
		case "retried":
			result.Retried++
		}
	}
	return result, nil
}

func (h *ProcessDueHandler) attemptOne(ctx context.Context, d *domain.Delivery) (string, error) {
	now := h.now().UTC()
	res, deliverErr := h.deliverer.Deliver(ctx, d.ID().String(), d.EventName(), d.URL(), d.Payload())

	var outcome string
	err := h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		current, err := h.repo.FindByID(ctx, d.ID())
		if err != nil {
			return err
		}
		if current.Status() != domain.StatusPending {
			outcome = "skipped"
			return nil
		}

		if deliverErr == nil && res.StatusCode >= 200 && res.StatusCode < 300 {
			if err := current.MarkSent(res.StatusCode, now); err != nil {
				return err
			}
			if err := h.repo.Save(ctx, current); err != nil {
				return fmt.Errorf("save delivery: %w", err)
			}
			if err := h.pub.Publish(ctx, "webhook_delivery", current.ID().String(), deliverySucceededEvent{
				DeliveryID:  current.ID().String(),
				MerchantID:  current.MerchantID(),
				SourceEvent: current.EventName(),
				SourceID:    current.SourceID(),
				OccurredAt:  now,
			}); err != nil {
				return err
			}
			outcome = "sent"
			return nil
		}

		errMsg := res.Error
		if deliverErr != nil {
			errMsg = deliverErr.Error()
		}
		if errMsg == "" {
			errMsg = fmt.Sprintf("http status %d", res.StatusCode)
		}
		retryable := res.Retryable
		if deliverErr != nil {
			retryable = true
		}

		if err := current.MarkAttemptFailed(res.StatusCode, errMsg, retryable, now); err != nil {
			return err
		}
		if err := h.repo.Save(ctx, current); err != nil {
			return fmt.Errorf("save delivery: %w", err)
		}

		if current.Status() == domain.StatusFailed {
			if err := h.pub.Publish(ctx, "webhook_delivery", current.ID().String(), deliveryFailedEvent{
				DeliveryID:  current.ID().String(),
				MerchantID:  current.MerchantID(),
				SourceEvent: current.EventName(),
				SourceID:    current.SourceID(),
				Attempt:     current.AttemptCount(),
				OccurredAt:  now,
			}); err != nil {
				return err
			}
			outcome = "failed"
			return nil
		}
		outcome = "retried"
		return nil
	})
	return outcome, err
}
