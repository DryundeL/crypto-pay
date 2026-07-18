package eventbus

// IntegrationEvent is a cross-module contract payload.
// Domain events stay inside module/internal/domain.
// Integration events cross BC boundaries via transactional outbox (+ optional Watermill relay).
type IntegrationEvent interface {
	EventName() string
}
