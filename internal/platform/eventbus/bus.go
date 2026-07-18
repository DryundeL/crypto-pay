package eventbus

import "context"

// Publisher publishes already-persisted outbox payloads to an async bus.
// Not used for synchronous command/query dispatch.
type Publisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

// Noop is a placeholder until Watermill (or another bus) is wired in the worker.
type Noop struct{}

func (Noop) Publish(context.Context, string, []byte) error { return nil }
