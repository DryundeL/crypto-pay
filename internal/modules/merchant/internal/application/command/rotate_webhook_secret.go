package command

type RotateWebhookSecret struct {
	MerchantID string
}

type RotateWebhookSecretResult struct {
	MerchantID    string
	WebhookSecret string // plaintext — returned once
}
