package ports

import (
	"context"

	"github.com/DryundeL/crypto-pay/internal/platform/eventbus"
)

// EventPublisher appends integration events to the transactional outbox.
type EventPublisher interface {
	Publish(ctx context.Context, aggregateType, aggregateID string, event eventbus.IntegrationEvent) error
}
