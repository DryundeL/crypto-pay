package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/domain"
)

type CreateAPIKeyHandler struct {
	repo      domain.Repository
	tx        ports.TransactionManager
	publisher ports.EventPublisher
	keys      ports.APIKeyGenerator
	now       func() time.Time
}

func NewCreateAPIKeyHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	publisher ports.EventPublisher,
	keys ports.APIKeyGenerator,
) *CreateAPIKeyHandler {
	return &CreateAPIKeyHandler{
		repo:      repo,
		tx:        tx,
		publisher: publisher,
		keys:      keys,
		now:       time.Now,
	}
}

type CreateAPIKeyResult struct {
	MerchantID string
	APIKeyID   string
	Name       string
	KeyPrefix  string
	APIKey     string // plaintext — returned once
	CreatedAt  time.Time
}

func (h *CreateAPIKeyHandler) Handle(ctx context.Context, cmd CreateAPIKey) (CreateAPIKeyResult, error) {
	var result CreateAPIKeyResult

	err := h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		m, err := h.repo.FindByID(ctx, domain.MerchantID(cmd.MerchantID))
		if err != nil {
			return err
		}

		material, err := h.keys.Generate()
		if err != nil {
			return fmt.Errorf("generate api key: %w", err)
		}

		now := h.now().UTC()
		key, err := m.IssueAPIKey(domain.IssueAPIKeyParams{
			ID:        domain.APIKeyID(uuid.NewString()),
			Name:      cmd.Name,
			KeyPrefix: material.Prefix,
			KeyHash:   material.Hash,
			Now:       now,
		})
		if err != nil {
			return err
		}

		if err := h.repo.Save(ctx, m); err != nil {
			return fmt.Errorf("save merchant: %w", err)
		}

		event := merchantAPIKeyCreatedEvent{
			MerchantID: m.ID().String(),
			APIKeyID:   key.ID().String(),
			KeyPrefix:  key.KeyPrefix(),
			OccurredAt: now,
		}
		if err := h.publisher.Publish(ctx, "merchant", m.ID().String(), event); err != nil {
			return err
		}

		result = CreateAPIKeyResult{
			MerchantID: m.ID().String(),
			APIKeyID:   key.ID().String(),
			Name:       key.Name(),
			KeyPrefix:  key.KeyPrefix(),
			APIKey:     material.Plaintext,
			CreatedAt:  key.CreatedAt(),
		}
		return nil
	})
	if err != nil {
		return CreateAPIKeyResult{}, err
	}
	return result, nil
}
