package command

import (
	"context"
	"fmt"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/domain"
)

type RevokeAPIKeyHandler struct {
	repo      domain.Repository
	tx        ports.TransactionManager
	publisher ports.EventPublisher
	now       func() time.Time
}

func NewRevokeAPIKeyHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	publisher ports.EventPublisher,
) *RevokeAPIKeyHandler {
	return &RevokeAPIKeyHandler{
		repo:      repo,
		tx:        tx,
		publisher: publisher,
		now:       time.Now,
	}
}

func (h *RevokeAPIKeyHandler) Handle(ctx context.Context, cmd RevokeAPIKey) error {
	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		m, err := h.repo.FindByID(ctx, domain.MerchantID(cmd.MerchantID))
		if err != nil {
			return err
		}

		now := h.now().UTC()
		key, err := m.RevokeAPIKey(domain.APIKeyID(cmd.APIKeyID), now)
		if err != nil {
			return err
		}

		if err := h.repo.Save(ctx, m); err != nil {
			return fmt.Errorf("save merchant: %w", err)
		}

		event := merchantAPIKeyRevokedEvent{
			MerchantID: m.ID().String(),
			APIKeyID:   key.ID().String(),
			OccurredAt: now,
		}
		return h.publisher.Publish(ctx, "merchant", m.ID().String(), event)
	})
}
