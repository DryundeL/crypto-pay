package ports

import "context"

// MerchantEndpoint resolves the merchant webhook URL.
type MerchantEndpoint interface {
	WebhookURL(ctx context.Context, merchantID string) (string, error)
}
