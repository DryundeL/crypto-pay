package http

type createMerchantRequest struct {
	Name       string `json:"name"`
	WebhookURL string `json:"webhook_url"`
}

type createAPIKeyRequest struct {
	Name string `json:"name"`
}
