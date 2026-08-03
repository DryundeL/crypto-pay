package webhook

import (
	"errors"
	"time"
)

// Public contracts of the Webhook bounded context.

type DeliveryStatus string

const (
	DeliveryPending DeliveryStatus = "pending"
	DeliverySent    DeliveryStatus = "sent"
	DeliveryFailed  DeliveryStatus = "failed"
)

var (
	ErrNotFound     = errors.New("webhook delivery not found")
	ErrInvalidInput = errors.New("invalid webhook delivery input")
)

// EnqueueInput creates a merchant webhook delivery from a business event.
type EnqueueInput struct {
	MerchantID string
	EventName  string
	SourceID   string
	Data       map[string]any
	OccurredAt time.Time
}

type EnqueueResult struct {
	ID      string
	Created bool
	Skipped bool
}

type ProcessDueResult struct {
	Processed int
	Sent      int
	Failed    int
	Retried   int
}
