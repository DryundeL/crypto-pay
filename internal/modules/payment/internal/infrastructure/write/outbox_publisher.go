package write

import (
	"context"
	"fmt"

	"github.com/DryundeL/crypto-pay/internal/platform/eventbus"
	"github.com/DryundeL/crypto-pay/internal/platform/outbox"
)

// OutboxPublisher implements application ports.EventPublisher via transactional outbox.
type OutboxPublisher struct {
	writer outbox.Writer
}

func NewOutboxPublisher(writer outbox.Writer) *OutboxPublisher {
	return &OutboxPublisher{writer: writer}
}

func (p *OutboxPublisher) Publish(
	ctx context.Context,
	aggregateType, aggregateID string,
	event eventbus.IntegrationEvent,
) error {
	msg, err := outbox.NewMessage(event.EventName(), aggregateType, aggregateID, event)
	if err != nil {
		return fmt.Errorf("build outbox message: %w", err)
	}
	if err := p.writer.Append(ctx, msg); err != nil {
		return fmt.Errorf("publish to outbox: %w", err)
	}
	return nil
}
