package command

import "time"

// Integration event payloads for outbox (mirror public merchant.events contracts).
// Kept here to avoid import cycle: merchant -> command -> merchant.

type merchantCreatedEvent struct {
	MerchantID string    `json:"merchant_id"`
	Name       string    `json:"name"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (merchantCreatedEvent) EventName() string { return "merchant.created" }

type merchantAPIKeyCreatedEvent struct {
	MerchantID string    `json:"merchant_id"`
	APIKeyID   string    `json:"api_key_id"`
	KeyPrefix  string    `json:"key_prefix"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (merchantAPIKeyCreatedEvent) EventName() string { return "merchant.api_key_created" }

type merchantAPIKeyRevokedEvent struct {
	MerchantID string    `json:"merchant_id"`
	APIKeyID   string    `json:"api_key_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (merchantAPIKeyRevokedEvent) EventName() string { return "merchant.api_key_revoked" }

type merchantWebhookSecretRotatedEvent struct {
	MerchantID string    `json:"merchant_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (merchantWebhookSecretRotatedEvent) EventName() string { return "merchant.webhook_secret_rotated" }
