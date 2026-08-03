package ledger

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/application/query"
	ledgerhttp "github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/delivery/http"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/infrastructure/read"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/infrastructure/write"
	"github.com/DryundeL/crypto-pay/internal/platform/outbox"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// Module is the Ledger bounded context composition root.
type Module struct {
	handler          *ledgerhttp.Handler
	postJournal      *command.PostJournalHandler
	postDebitJournal *command.PostDebitJournalHandler
}

type Dependencies struct {
	DB *gorm.DB
}

func NewModule(deps Dependencies) *Module {
	if deps.DB == nil {
		panic("ledger: DB is required")
	}

	tx := transaction.NewGormManager(deps.DB)
	outboxStore := outbox.NewStore(deps.DB)
	publisher := write.NewOutboxPublisher(outboxStore)
	accounts := write.NewAccountRepository(deps.DB)
	journals := write.NewJournalRepository(deps.DB)
	queries := read.NewLedgerQueries(deps.DB)

	postJournal := command.NewPostJournalHandler(accounts, journals, tx, publisher)
	postDebitJournal := command.NewPostDebitJournalHandler(accounts, journals, tx, publisher)
	listBalances := query.NewListBalancesHandler(queries)
	listJournals := query.NewListJournalsHandler(queries)

	return &Module{
		handler:          ledgerhttp.NewHandler(listBalances, listJournals),
		postJournal:      postJournal,
		postDebitJournal: postDebitJournal,
	}
}

func (m *Module) Name() string { return "ledger" }

func (m *Module) RegisterHTTP(g *echo.Group, authMiddleware echo.MiddlewareFunc) {
	ledgerhttp.RegisterRoutes(g, ledgerhttp.RouteDeps{
		Handler:        m.handler,
		AuthMiddleware: authMiddleware,
	})
}

// PostJournal posts a balanced credit journal (debit clearing / credit merchant available).
// Idempotent on IdempotencyKey — retries return the existing journal.
func (m *Module) PostJournal(ctx context.Context, in PostJournalInput) (PostJournalResult, error) {
	result, err := m.postJournal.Handle(ctx, command.PostJournal{
		IdempotencyKey: in.IdempotencyKey,
		MerchantID:     in.MerchantID,
		Amount:         in.Amount,
		Currency:       in.Currency,
		ReferenceType:  in.ReferenceType,
		ReferenceID:    in.ReferenceID,
	})
	if err != nil {
		return PostJournalResult{}, mapPublicError(err)
	}
	return PostJournalResult{
		JournalID: result.JournalID,
		Created:   result.Created,
	}, nil
}

// PostDebitJournal posts a balanced debit journal (debit merchant available / credit clearing).
// Idempotent on IdempotencyKey — retries return the existing journal.
func (m *Module) PostDebitJournal(ctx context.Context, in PostDebitJournalInput) (PostDebitJournalResult, error) {
	result, err := m.postDebitJournal.Handle(ctx, command.PostDebitJournal{
		IdempotencyKey: in.IdempotencyKey,
		MerchantID:     in.MerchantID,
		Amount:         in.Amount,
		Currency:       in.Currency,
		ReferenceType:  in.ReferenceType,
		ReferenceID:    in.ReferenceID,
	})
	if err != nil {
		return PostDebitJournalResult{}, mapPublicError(err)
	}
	return PostDebitJournalResult{
		JournalID: result.JournalID,
		Created:   result.Created,
	}, nil
}

func mapPublicError(err error) error {
	switch {
	case errors.Is(err, domain.ErrJournalNotFound),
		errors.Is(err, domain.ErrAccountNotFound):
		return ErrNotFound
	case errors.Is(err, domain.ErrInsufficientBalance):
		return ErrInsufficientBalance
	case errors.Is(err, domain.ErrInvalidMoney),
		errors.Is(err, domain.ErrInvalidJournal),
		errors.Is(err, domain.ErrInvalidAccount):
		return errors.Join(ErrInvalidInput, err)
	default:
		return err
	}
}
