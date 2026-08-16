package eventbus

import (
	"bytes"
	"context"
	"slices"
	"sync"
)

// Handler consumes a published outbox payload. Must be idempotent:
// the relayer delivers at-least-once (retry after a handler/publish error).
type Handler func(ctx context.Context, topic string, payload []byte) error

// InProcess is a process-local bus for the worker. Watermill can replace it later
// without changing Publisher or outbox.Relayer.
type InProcess struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewInProcess() *InProcess {
	return &InProcess{handlers: make(map[string][]Handler)}
}

func (b *InProcess) Subscribe(topic string, h Handler) {
	if topic == "" || h == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[topic] = append(b.handlers[topic], h)
}

func (b *InProcess) Publish(ctx context.Context, topic string, payload []byte) error {
	b.mu.RLock()
	handlers := slices.Clone(b.handlers[topic])
	b.mu.RUnlock()

	body := bytes.Clone(payload)
	for _, h := range handlers {
		if err := h(ctx, topic, body); err != nil {
			return err
		}
	}
	return nil
}
