package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/domain"
)

type CreateMerchantHandler struct {
	repo      domain.Repository
	tx        ports.TransactionManager
	publisher ports.EventPublisher
	keys      ports.APIKeyGenerator
	now       func() time.Time
	newSecret func() (string, error)
}

func NewCreateMerchantHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	publisher ports.EventPublisher,
	keys ports.APIKeyGenerator,
) *CreateMerchantHandler {
	return &CreateMerchantHandler{
		repo:      repo,
		tx:        tx,
		publisher: publisher,
		keys:      keys,
		now:       time.Now,
		newSecret: domain.GenerateWebhookSecret,
	}
}

type CreateMerchantResult struct {
	MerchantID    string
	Name          string
	Status        string
	WebhookURL    string
	CreatedAt     time.Time
	APIKeyID      string
	KeyPrefix     string
	APIKey        string // plaintext — returned once (bootstrap key)
	WebhookSecret string // plaintext — returned once
}

func (h *CreateMerchantHandler) Handle(ctx context.Context, cmd CreateMerchant) (CreateMerchantResult, error) {
	var result CreateMerchantResult

	err := h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		now := h.now().UTC()
		secret, err := h.newSecret()
		if err != nil {
			return err
		}
		m, err := domain.NewMerchant(domain.NewMerchantParams{
			ID:            domain.MerchantID(uuid.NewString()),
			Name:          cmd.Name,
			WebhookURL:    cmd.WebhookURL,
			WebhookSecret: secret,
			Now:           now,
		})
		if err != nil {
			return err
		}

		material, err := h.keys.Generate()
		if err != nil {
			return fmt.Errorf("generate api key: %w", err)
		}

		key, err := m.IssueAPIKey(domain.IssueAPIKeyParams{
			ID:        domain.APIKeyID(uuid.NewString()),
			Name:      "default",
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

		if err := h.publisher.Publish(ctx, "merchant", m.ID().String(), merchantCreatedEvent{
			MerchantID: m.ID().String(),
			Name:       m.Name(),
			OccurredAt: now,
		}); err != nil {
			return err
		}

		if err := h.publisher.Publish(ctx, "merchant", m.ID().String(), merchantAPIKeyCreatedEvent{
			MerchantID: m.ID().String(),
			APIKeyID:   key.ID().String(),
			KeyPrefix:  key.KeyPrefix(),
			OccurredAt: now,
		}); err != nil {
			return err
		}

		result = CreateMerchantResult{
			MerchantID:    m.ID().String(),
			Name:          m.Name(),
			Status:        m.Status().String(),
			WebhookURL:    m.WebhookURL(),
			CreatedAt:     m.CreatedAt(),
			APIKeyID:      key.ID().String(),
			KeyPrefix:     key.KeyPrefix(),
			APIKey:        material.Plaintext,
			WebhookSecret: m.WebhookSecret(),
		}
		return nil
	})
	if err != nil {
		return CreateMerchantResult{}, err
	}
	return result, nil
}
