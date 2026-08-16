package ports

import "context"

// MerchantEndpoint resolves the merchant webhook URL and HMAC secret.
type MerchantEndpoint interface {
	WebhookURL(ctx context.Context, merchantID string) (string, error)
	WebhookSecret(ctx context.Context, merchantID string) (string, error)
}
