package invoice

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/query"
	invoicehttp "github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/delivery/http"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/infrastructure/read"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/infrastructure/write"
	"github.com/DryundeL/crypto-pay/internal/platform/outbox"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// Module is the Invoice bounded context composition root (within the module).
type Module struct {
	handler *invoicehttp.Handler

	findPendingByAddress *query.FindPendingByAddressHandler
	markConfirming       *command.MarkConfirmingHandler
	markPaid             *command.MarkPaidHandler
	paidNotifier         PaidNotifier
	expireDue            *command.ExpireDueHandler
}

type Dependencies struct {
	DB              *gorm.DB
	AddressProvider ports.AddressProvider
}

func NewModule(deps Dependencies) *Module {
	if deps.DB == nil {
		panic("invoice: DB is required")
	}
	if deps.AddressProvider == nil {
		panic("invoice: AddressProvider is required")
	}

	tx := transaction.NewGormManager(deps.DB)
	outboxStore := outbox.NewStore(deps.DB)
	publisher := write.NewOutboxPublisher(outboxStore)
	repo := write.NewInvoiceRepository(deps.DB)
	queries := read.NewInvoiceQueries(deps.DB)

	createInvoice := command.NewCreateInvoiceHandler(repo, tx, publisher, deps.AddressProvider)
	cancelInvoice := command.NewCancelInvoiceHandler(repo, tx)
	expireInvoice := command.NewExpireInvoiceHandler(repo, tx, publisher)
	expireDue := command.NewExpireDueHandler(repo, expireInvoice)
	markConfirming := command.NewMarkConfirmingHandler(repo, tx)
	markPaid := command.NewMarkPaidHandler(repo, tx, publisher)
	getInvoice := query.NewGetInvoiceHandler(queries)
	findPending := query.NewFindPendingByAddressHandler(queries)

	return &Module{
		handler:              invoicehttp.NewHandler(createInvoice, cancelInvoice, getInvoice),
		findPendingByAddress: findPending,
		markConfirming:       markConfirming,
		markPaid:             markPaid,
		expireDue:            expireDue,
	}
}

func (m *Module) Name() string { return "invoice" }

func (m *Module) RegisterHTTP(g *echo.Group, authMiddleware echo.MiddlewareFunc) {
	invoicehttp.RegisterRoutes(g, invoicehttp.RouteDeps{
		Handler:        m.handler,
		AuthMiddleware: authMiddleware,
	})
}

// FindPendingByAddress looks up a pending invoice by deposit network+address.
func (m *Module) FindPendingByAddress(ctx context.Context, network, address string) (InvoiceRef, error) {
	dto, err := m.findPendingByAddress.Handle(ctx, query.FindPendingByAddress{
		Network: network,
		Address: address,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return InvoiceRef{}, ErrNotFound
		}
		return InvoiceRef{}, err
	}
	return InvoiceRef{
		ID:         dto.ID,
		MerchantID: dto.MerchantID,
		Amount:     dto.Amount,
		Currency:   dto.Currency,
		Network:    dto.Network,
		Address:    dto.Address,
		Status:     dto.Status,
	}, nil
}

// MarkConfirming transitions a pending invoice to confirming.
func (m *Module) MarkConfirming(ctx context.Context, invoiceID string) error {
	err := m.markConfirming.Handle(ctx, command.MarkConfirming{InvoiceID: invoiceID})
	if errors.Is(err, domain.ErrInvoiceNotFound) {
		return ErrNotFound
	}
	return err
}

// SetPaidNotifier wires a post-paid side effect (e.g. webhook enqueue) at composition root.
func (m *Module) SetPaidNotifier(n PaidNotifier) {
	m.paidNotifier = n
}

// MarkPaid transitions invoice to paid and emits invoice.paid via outbox.
func (m *Module) MarkPaid(ctx context.Context, invoiceID, txHash string) error {
	result, err := m.markPaid.Handle(ctx, command.MarkPaid{InvoiceID: invoiceID, TxHash: txHash})
	if errors.Is(err, domain.ErrInvoiceNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if m.paidNotifier != nil {
		if err := m.paidNotifier.NotifyInvoicePaid(ctx, InvoicePaidNotification{
			InvoiceID:  result.InvoiceID,
			MerchantID: result.MerchantID,
			Amount:     result.Amount,
			Currency:   result.Currency,
			TxHash:     result.TxHash,
			OccurredAt: result.OccurredAt,
		}); err != nil {
			return err
		}
	}
	return nil
}

// ExpireDue expires pending invoices whose TTL has elapsed (sync facade for worker).
func (m *Module) ExpireDue(ctx context.Context, limit int) (ExpireDueResult, error) {
	result, err := m.expireDue.Handle(ctx, command.ExpireDue{Limit: limit})
	if err != nil {
		return ExpireDueResult{}, err
	}
	return ExpireDueResult{
		Expired: result.Expired,
		Skipped: result.Skipped,
	}, nil
}
