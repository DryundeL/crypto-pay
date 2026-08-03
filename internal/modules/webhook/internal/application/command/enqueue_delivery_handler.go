package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/domain"
)

type EnqueueDeliveryHandler struct {
	repo     domain.Repository
	tx       ports.TransactionManager
	merchant ports.MerchantEndpoint
	now      func() time.Time
	newID    func() string
}

func NewEnqueueDeliveryHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	merchant ports.MerchantEndpoint,
) *EnqueueDeliveryHandler {
	return &EnqueueDeliveryHandler{
		repo:     repo,
		tx:       tx,
		merchant: merchant,
		now:      time.Now,
		newID:    func() string { return uuid.NewString() },
	}
}

func (h *EnqueueDeliveryHandler) Handle(ctx context.Context, cmd EnqueueDelivery) (EnqueueDeliveryResult, error) {
	if cmd.MerchantID == "" || cmd.EventName == "" || cmd.SourceID == "" {
		return EnqueueDeliveryResult{}, fmt.Errorf("%w: merchant_id, event_name and source_id are required", domain.ErrInvalidDelivery)
	}

	key := cmd.EventName + ":" + cmd.SourceID
	if existing, err := h.repo.FindByIdempotencyKey(ctx, key); err == nil {
		return EnqueueDeliveryResult{
			ID:        existing.ID().String(),
			Created:   false,
			CreatedAt: existing.CreatedAt(),
		}, nil
	} else if !errors.Is(err, domain.ErrDeliveryNotFound) {
		return EnqueueDeliveryResult{}, err
	}

	url, err := h.merchant.WebhookURL(ctx, cmd.MerchantID)
	if err != nil {
		return EnqueueDeliveryResult{}, err
	}
	if url == "" {
		return EnqueueDeliveryResult{Skipped: true}, nil
	}

	now := h.now().UTC()
	occurredAt := cmd.OccurredAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = now
	}

	id := h.newID()
	body, err := json.Marshal(map[string]any{
		"id":          id,
		"event":       cmd.EventName,
		"occurred_at": occurredAt,
		"data":        cmd.Data,
	})
	if err != nil {
		return EnqueueDeliveryResult{}, fmt.Errorf("marshal payload: %w", err)
	}

	d, err := domain.Enqueue(domain.EnqueueParams{
		ID:          domain.DeliveryID(id),
		MerchantID:  cmd.MerchantID,
		EventName:   cmd.EventName,
		SourceID:    cmd.SourceID,
		URL:         url,
		Payload:     body,
		MaxAttempts: domain.DefaultMaxAttempts,
		Now:         now,
	})
	if err != nil {
		return EnqueueDeliveryResult{}, err
	}

	created := true
	err = h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if again, err := h.repo.FindByIdempotencyKey(ctx, key); err == nil {
			d = again
			created = false
			return nil
		} else if !errors.Is(err, domain.ErrDeliveryNotFound) {
			return err
		}
		return h.repo.Save(ctx, d)
	})
	if err != nil {
		return EnqueueDeliveryResult{}, err
	}
	return EnqueueDeliveryResult{
		ID:        d.ID().String(),
		Created:   created,
		CreatedAt: d.CreatedAt(),
	}, nil
}
