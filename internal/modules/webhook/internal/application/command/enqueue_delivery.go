package command

import "time"

type EnqueueDelivery struct {
	MerchantID string
	EventName  string
	SourceID   string
	Data       map[string]any
	OccurredAt time.Time
}

type EnqueueDeliveryResult struct {
	ID        string
	Created   bool
	Skipped   bool // no webhook URL configured
	CreatedAt time.Time
}
