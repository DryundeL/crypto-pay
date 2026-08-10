package payment

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/application/query"
	paymenthttp "github.com/DryundeL/crypto-pay/internal/modules/payment/internal/delivery/http"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/infrastructure/read"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/infrastructure/write"
	"github.com/DryundeL/crypto-pay/internal/platform/outbox"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// Module is the Payment bounded context composition root (within the module).
type Module struct {
	handler        *paymenthttp.Handler
	recordObserved *command.RecordObservedHandler
	confirm        *command.ConfirmPaymentHandler
}

type Dependencies struct {
	DB               *gorm.DB
	InvoiceLookup    ports.InvoiceLookup
	InvoiceLifecycle ports.InvoiceLifecycle
	Confirmations    ConfirmationPolicy
}

func NewModule(deps Dependencies) *Module {
	if deps.DB == nil {
		panic("payment: DB is required")
	}
	if deps.InvoiceLookup == nil || deps.InvoiceLifecycle == nil {
		panic("payment: invoice ports are required")
	}
	if !deps.Confirmations.Configured() {
		panic("payment: Confirmations policy is required")
	}

	tx := transaction.NewGormManager(deps.DB)
	outboxStore := outbox.NewStore(deps.DB)
	publisher := write.NewOutboxPublisher(outboxStore)
	repo := write.NewPaymentRepository(deps.DB)
	queries := read.NewPaymentQueries(deps.DB)

	recordObserved := command.NewRecordObservedHandler(
		repo, tx, publisher, deps.InvoiceLookup, deps.InvoiceLifecycle,
	)
	confirm := command.NewConfirmPaymentHandler(
		repo, tx, publisher, deps.InvoiceLifecycle, deps.Confirmations,
	)
	getPayment := query.NewGetPaymentHandler(queries)

	return &Module{
		handler:        paymenthttp.NewHandler(getPayment),
		recordObserved: recordObserved,
		confirm:        confirm,
	}
}

func (m *Module) Name() string { return "payment" }

func (m *Module) RegisterHTTP(g *echo.Group, authMiddleware echo.MiddlewareFunc) {
	paymenthttp.RegisterRoutes(g, paymenthttp.RouteDeps{
		Handler:        m.handler,
		AuthMiddleware: authMiddleware,
	})
}

// RecordObserved records a detected on-chain payment (sync facade for blockchain).
func (m *Module) RecordObserved(ctx context.Context, in RecordObservedInput) (PaymentView, error) {
	result, err := m.recordObserved.Handle(ctx, command.RecordObserved{
		MerchantID:    in.MerchantID,
		Network:       in.Network,
		TxHash:        in.TxHash,
		ToAddress:     in.ToAddress,
		Amount:        in.Amount,
		Currency:      in.Currency,
		Confirmations: in.Confirmations,
	})
	if err != nil {
		return PaymentView{}, mapFacadeError(err)
	}
	return PaymentView{
		ID:            result.ID,
		InvoiceID:     result.InvoiceID,
		MerchantID:    result.MerchantID,
		Status:        result.Status,
		Network:       result.Network,
		TxHash:        result.TxHash,
		ToAddress:     result.ToAddress,
		Amount:        result.Amount,
		Currency:      result.Currency,
		Confirmations: result.Confirmations,
		CreatedAt:     result.CreatedAt,
		UpdatedAt:     result.UpdatedAt,
	}, nil
}

// Confirm confirms a previously observed payment (sync facade for blockchain).
func (m *Module) Confirm(ctx context.Context, in ConfirmInput) (PaymentView, error) {
	result, err := m.confirm.Handle(ctx, command.ConfirmPayment{
		MerchantID: in.MerchantID,
		Network:    in.Network,
		TxHash:     in.TxHash,
	})
	if err != nil {
		return PaymentView{}, mapFacadeError(err)
	}
	return PaymentView{
		ID:            result.ID,
		InvoiceID:     result.InvoiceID,
		MerchantID:    result.MerchantID,
		Status:        result.Status,
		Network:       result.Network,
		TxHash:        result.TxHash,
		ToAddress:     result.ToAddress,
		Amount:        result.Amount,
		Currency:      result.Currency,
		Confirmations: result.Confirmations,
		CreatedAt:     result.CreatedAt,
		UpdatedAt:     result.UpdatedAt,
	}, nil
}

func mapFacadeError(err error) error {
	switch {
	case errors.Is(err, domain.ErrPaymentNotFound):
		return ErrNotFound
	case errors.Is(err, domain.ErrInsufficientConfirmations):
		return ErrInsufficientConfirmations
	case errors.Is(err, ErrInvalid),
		errors.Is(err, domain.ErrInvalidPayment),
		errors.Is(err, domain.ErrInvalidTransition):
		return ErrInvalid
	default:
		return err
	}
}
