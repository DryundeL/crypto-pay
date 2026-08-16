package command

import (
	"context"
	"fmt"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/domain"
)

type RotateWebhookSecretHandler struct {
	repo      domain.Repository
	tx        ports.TransactionManager
	publisher ports.EventPublisher
	now       func() time.Time
}

func NewRotateWebhookSecretHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	publisher ports.EventPublisher,
) *RotateWebhookSecretHandler {
	return &RotateWebhookSecretHandler{
		repo:      repo,
		tx:        tx,
		publisher: publisher,
		now:       time.Now,
	}
}

func (h *RotateWebhookSecretHandler) Handle(ctx context.Context, cmd RotateWebhookSecret) (RotateWebhookSecretResult, error) {
	var result RotateWebhookSecretResult

	err := h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		m, err := h.repo.FindByID(ctx, domain.MerchantID(cmd.MerchantID))
		if err != nil {
			return err
		}

		secret, err := domain.GenerateWebhookSecret()
		if err != nil {
			return err
		}

		now := h.now().UTC()
		if err := m.RotateWebhookSecret(secret, now); err != nil {
			return err
		}
		if err := h.repo.Save(ctx, m); err != nil {
			return fmt.Errorf("save merchant: %w", err)
		}
		if err := h.publisher.Publish(ctx, "merchant", m.ID().String(), merchantWebhookSecretRotatedEvent{
			MerchantID: m.ID().String(),
			OccurredAt: now,
		}); err != nil {
			return err
		}

		result = RotateWebhookSecretResult{
			MerchantID:    m.ID().String(),
			WebhookSecret: secret,
		}
		return nil
	})
	if err != nil {
		return RotateWebhookSecretResult{}, err
	}
	return result, nil
}
