package invoice

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/query"
	invoicehttp "github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/delivery/http"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/infrastructure/read"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/infrastructure/write"
	"github.com/DryundeL/crypto-pay/internal/platform/outbox"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// Module is the Invoice bounded context composition root (within the module).
type Module struct {
	handler *invoicehttp.Handler

	// Internal commands for other processes / future event handlers (no HTTP).
	ExpireInvoice  *command.ExpireInvoiceHandler
	MarkConfirming *command.MarkConfirmingHandler
	MarkPaid       *command.MarkPaidHandler
}

type Dependencies struct {
	DB *gorm.DB
}

func NewModule(deps Dependencies) *Module {
	tx := transaction.NewGormManager(deps.DB)
	outboxStore := outbox.NewStore(deps.DB)
	publisher := write.NewOutboxPublisher(outboxStore)
	repo := write.NewInvoiceRepository(deps.DB)
	queries := read.NewInvoiceQueries(deps.DB)
	addresses := write.NewStubAddressProvider()

	createInvoice := command.NewCreateInvoiceHandler(repo, tx, publisher, addresses)
	cancelInvoice := command.NewCancelInvoiceHandler(repo, tx)
	expireInvoice := command.NewExpireInvoiceHandler(repo, tx, publisher)
	markConfirming := command.NewMarkConfirmingHandler(repo, tx)
	markPaid := command.NewMarkPaidHandler(repo, tx, publisher)
	getInvoice := query.NewGetInvoiceHandler(queries)

	return &Module{
		handler:        invoicehttp.NewHandler(createInvoice, cancelInvoice, getInvoice),
		ExpireInvoice:  expireInvoice,
		MarkConfirming: markConfirming,
		MarkPaid:       markPaid,
	}
}

func (m *Module) Name() string { return "invoice" }

func (m *Module) RegisterHTTP(g *echo.Group, authMiddleware echo.MiddlewareFunc) {
	invoicehttp.RegisterRoutes(g, invoicehttp.RouteDeps{
		Handler:        m.handler,
		AuthMiddleware: authMiddleware,
	})
}
