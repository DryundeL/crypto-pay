package domain

import "time"

// DomainEvent is a marker for invoice domain events.
type DomainEvent interface {
	OccurredAt() time.Time
	EventName() string
}
