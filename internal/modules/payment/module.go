package payment

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/application/query"
	paymenthttp "github.com/DryundeL/crypto-pay/internal/modules/payment/internal/delivery/http"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/infrastructure/read"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/infrastructure/write"
	"github.com/DryundeL/crypto-pay/internal/platform/outbox"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// Module is the Payment bounded context composition root (within the module).
type Module struct {
	handler *paymenthttp.Handler
}

type Dependencies struct {
	DB               *gorm.DB
	InvoiceLookup    ports.InvoiceLookup
	InvoiceLifecycle ports.InvoiceLifecycle
}

func NewModule(deps Dependencies) *Module {
	tx := transaction.NewGormManager(deps.DB)
	outboxStore := outbox.NewStore(deps.DB)
	publisher := write.NewOutboxPublisher(outboxStore)
	repo := write.NewPaymentRepository(deps.DB)
	queries := read.NewPaymentQueries(deps.DB)

	recordObserved := command.NewRecordObservedHandler(
		repo, tx, publisher, deps.InvoiceLookup, deps.InvoiceLifecycle,
	)
	confirm := command.NewConfirmPaymentHandler(repo, tx, publisher, deps.InvoiceLifecycle)
	getPayment := query.NewGetPaymentHandler(queries)

	return &Module{
		handler: paymenthttp.NewHandler(recordObserved, confirm, getPayment),
	}
}

func (m *Module) Name() string { return "payment" }

func (m *Module) RegisterHTTP(g *echo.Group, authMiddleware echo.MiddlewareFunc) {
	paymenthttp.RegisterRoutes(g, paymenthttp.RouteDeps{
		Handler:        m.handler,
		AuthMiddleware: authMiddleware,
	})
}
