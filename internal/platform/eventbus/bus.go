package eventbus

import "context"

// Publisher publishes already-persisted outbox payloads to an async bus.
// Not used for synchronous command/query dispatch.
type Publisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

// Noop drops messages. Production worker uses InProcess (Watermill optional later).
type Noop struct{}

func (Noop) Publish(context.Context, string, []byte) error { return nil }
