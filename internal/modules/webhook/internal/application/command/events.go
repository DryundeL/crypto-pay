package command

import "time"

type deliverySucceededEvent struct {
	DeliveryID  string    `json:"delivery_id"`
	MerchantID  string    `json:"merchant_id"`
	SourceEvent string    `json:"event_name"`
	SourceID    string    `json:"source_id"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func (deliverySucceededEvent) EventName() string { return "webhook.delivery_succeeded" }

type deliveryFailedEvent struct {
	DeliveryID  string    `json:"delivery_id"`
	MerchantID  string    `json:"merchant_id"`
	SourceEvent string    `json:"event_name"`
	SourceID    string    `json:"source_id"`
	Attempt     int       `json:"attempt"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func (deliveryFailedEvent) EventName() string { return "webhook.delivery_failed" }
