package outbox

import "context"

// Writer appends integration events inside the caller's DB transaction.
// Command handlers must persist aggregate changes and outbox messages atomically.
type Writer interface {
	Append(ctx context.Context, msg Message) error
}

// Relayer publishes pending outbox messages to an event bus (in-process or Watermill).
// Runs in worker process — not on the synchronous command path.
type Relayer interface {
	RelayPending(ctx context.Context, limit int) (published int, err error)
}
