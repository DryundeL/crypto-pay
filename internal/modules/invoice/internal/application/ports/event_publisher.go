package ports

import (
	"context"

	"github.com/DryundeL/crypto-pay/internal/platform/eventbus"
)

// EventPublisher appends integration events to the transactional outbox.
// Must be invoked inside the same DB transaction as aggregate persistence.
// Does not talk to Watermill/broker synchronously.
type EventPublisher interface {
	Publish(ctx context.Context, aggregateType, aggregateID string, event eventbus.IntegrationEvent) error
}
