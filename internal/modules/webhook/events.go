package webhook

import "time"

const (
	EventWebhookDeliverySucceeded = "webhook.delivery_succeeded"
	EventWebhookDeliveryFailed    = "webhook.delivery_failed"
)

type WebhookDeliverySucceeded struct {
	DeliveryID string    `json:"delivery_id"`
	InvoiceID  string    `json:"invoice_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (WebhookDeliverySucceeded) EventName() string { return EventWebhookDeliverySucceeded }

type WebhookDeliveryFailed struct {
	DeliveryID string    `json:"delivery_id"`
	InvoiceID  string    `json:"invoice_id"`
	Attempt    int       `json:"attempt"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (WebhookDeliveryFailed) EventName() string { return EventWebhookDeliveryFailed }
