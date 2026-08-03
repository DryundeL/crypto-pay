package query

import "time"

type DeliveryDTO struct {
	ID             string     `json:"id"`
	MerchantID     string     `json:"merchant_id"`
	EventName      string     `json:"event_name"`
	SourceID       string     `json:"source_id"`
	URL            string     `json:"url"`
	Status         string     `json:"status"`
	AttemptCount   int        `json:"attempt_count"`
	MaxAttempts    int        `json:"max_attempts"`
	NextAttemptAt  time.Time  `json:"next_attempt_at"`
	LastError      string     `json:"last_error,omitempty"`
	LastStatusCode *int       `json:"last_status_code,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
