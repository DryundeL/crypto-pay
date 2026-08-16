package merchant

import "time"

const (
	EventMerchantCreated              = "merchant.created"
	EventMerchantAPIKeyCreated        = "merchant.api_key_created"
	EventMerchantAPIKeyRevoked        = "merchant.api_key_revoked"
	EventMerchantWebhookSecretRotated = "merchant.webhook_secret_rotated"
)

type MerchantCreated struct {
	MerchantID string    `json:"merchant_id"`
	Name       string    `json:"name"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (MerchantCreated) EventName() string { return EventMerchantCreated }

type MerchantAPIKeyCreated struct {
	MerchantID string    `json:"merchant_id"`
	APIKeyID   string    `json:"api_key_id"`
	KeyPrefix  string    `json:"key_prefix"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (MerchantAPIKeyCreated) EventName() string { return EventMerchantAPIKeyCreated }

type MerchantAPIKeyRevoked struct {
	MerchantID string    `json:"merchant_id"`
	APIKeyID   string    `json:"api_key_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (MerchantAPIKeyRevoked) EventName() string { return EventMerchantAPIKeyRevoked }

type MerchantWebhookSecretRotated struct {
	MerchantID string    `json:"merchant_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (MerchantWebhookSecretRotated) EventName() string { return EventMerchantWebhookSecretRotated }
