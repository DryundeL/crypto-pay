package webhook

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/application/query"
	webhookhttp "github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/delivery/http"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/infrastructure/read"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/infrastructure/write"
	"github.com/DryundeL/crypto-pay/internal/platform/outbox"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// Module is the Webhook bounded context composition root.
type Module struct {
	handler *webhookhttp.Handler
	enqueue *command.EnqueueDeliveryHandler
	process *command.ProcessDueHandler
}

type Dependencies struct {
	DB       *gorm.DB
	Merchant ports.MerchantEndpoint
}

func NewModule(deps Dependencies) *Module {
	if deps.DB == nil {
		panic("webhook: DB is required")
	}
	if deps.Merchant == nil {
		panic("webhook: Merchant is required")
	}

	tx := transaction.NewGormManager(deps.DB)
	outboxStore := outbox.NewStore(deps.DB)
	publisher := write.NewOutboxPublisher(outboxStore)
	repo := write.NewDeliveryRepository(deps.DB)
	queries := read.NewDeliveryQueries(deps.DB)
	deliverer := write.NewHTTPDeliverer()

	enqueue := command.NewEnqueueDeliveryHandler(repo, tx, deps.Merchant)
	process := command.NewProcessDueHandler(repo, tx, publisher, deliverer, deps.Merchant)
	get := query.NewGetDeliveryHandler(queries)
	list := query.NewListDeliveriesHandler(queries)

	return &Module{
		handler: webhookhttp.NewHandler(get, list),
		enqueue: enqueue,
		process: process,
	}
}

func (m *Module) Name() string { return "webhook" }

func (m *Module) RegisterHTTP(g *echo.Group, authMiddleware echo.MiddlewareFunc) {
	webhookhttp.RegisterRoutes(g, webhookhttp.RouteDeps{
		Handler:        m.handler,
		AuthMiddleware: authMiddleware,
	})
}

// Enqueue creates a pending delivery for a merchant business event (sync facade).
func (m *Module) Enqueue(ctx context.Context, in EnqueueInput) (EnqueueResult, error) {
	result, err := m.enqueue.Handle(ctx, command.EnqueueDelivery{
		MerchantID: in.MerchantID,
		EventName:  in.EventName,
		SourceID:   in.SourceID,
		Data:       in.Data,
		OccurredAt: in.OccurredAt,
	})
	if err != nil {
		return EnqueueResult{}, mapPublicError(err)
	}
	return EnqueueResult{
		ID:      result.ID,
		Created: result.Created,
		Skipped: result.Skipped,
	}, nil
}

// ProcessDue attempts due pending deliveries (sync facade for worker).
func (m *Module) ProcessDue(ctx context.Context, limit int) (ProcessDueResult, error) {
	result, err := m.process.Handle(ctx, command.ProcessDue{Limit: limit})
	if err != nil {
		return ProcessDueResult{}, mapPublicError(err)
	}
	return ProcessDueResult{
		Processed: result.Processed,
		Sent:      result.Sent,
		Failed:    result.Failed,
		Retried:   result.Retried,
	}, nil
}

func mapPublicError(err error) error {
	switch {
	case errors.Is(err, domain.ErrDeliveryNotFound):
		return ErrNotFound
	case errors.Is(err, domain.ErrInvalidDelivery), errors.Is(err, domain.ErrInvalidTransition):
		return ErrInvalidInput
	default:
		return err
	}
}
