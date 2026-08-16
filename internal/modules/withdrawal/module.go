package withdrawal

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/application/query"
	withdrawalhttp "github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/delivery/http"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/infrastructure/read"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/infrastructure/write"
	"github.com/DryundeL/crypto-pay/internal/platform/outbox"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// Module is the Withdrawal bounded context composition root.
type Module struct {
	handler  *withdrawalhttp.Handler
	complete *command.CompleteWithdrawalHandler
}

type Dependencies struct {
	DB     *gorm.DB
	Ledger ports.LedgerDebiter
}

func NewModule(deps Dependencies) *Module {
	if deps.DB == nil {
		panic("withdrawal: DB is required")
	}
	if deps.Ledger == nil {
		panic("withdrawal: Ledger is required")
	}

	tx := transaction.NewGormManager(deps.DB)
	outboxStore := outbox.NewStore(deps.DB)
	publisher := write.NewOutboxPublisher(outboxStore)
	repo := write.NewWithdrawalRepository(deps.DB)
	queries := read.NewWithdrawalQueries(deps.DB)

	request := command.NewRequestWithdrawalHandler(repo, tx, publisher, deps.Ledger)
	complete := command.NewCompleteWithdrawalHandler(repo, tx, publisher)
	get := query.NewGetWithdrawalHandler(queries)
	list := query.NewListWithdrawalsHandler(queries)

	return &Module{
		handler:  withdrawalhttp.NewHandler(request, get, list),
		complete: complete,
	}
}

func (m *Module) Name() string { return "withdrawal" }

func (m *Module) RegisterHTTP(g *echo.Group, authMiddleware echo.MiddlewareFunc) {
	withdrawalhttp.RegisterRoutes(g, withdrawalhttp.RouteDeps{
		Handler:        m.handler,
		AuthMiddleware: authMiddleware,
	})
}

// Complete marks a requested withdrawal as completed (sync facade for worker).
// withdrawal.completed is persisted in the outbox in the same TX; the worker relays it.
func (m *Module) Complete(ctx context.Context, withdrawalID, txHash string) error {
	_, err := m.complete.Handle(ctx, command.CompleteWithdrawal{
		WithdrawalID: withdrawalID,
		TxHash:       txHash,
	})
	if err != nil {
		return mapPublicError(err)
	}
	return nil
}

func mapPublicError(err error) error {
	switch {
	case errors.Is(err, domain.ErrWithdrawalNotFound):
		return ErrNotFound
	case errors.Is(err, domain.ErrInvalidTransition):
		return ErrInvalidTransition
	case errors.Is(err, domain.ErrInvalidWithdrawal):
		return ErrInvalidInput
	case errors.Is(err, ledger.ErrInsufficientBalance):
		return ErrInsufficientBalance
	case errors.Is(err, ledger.ErrInvalidInput):
		return ErrInvalidInput
	default:
		return err
	}
}
