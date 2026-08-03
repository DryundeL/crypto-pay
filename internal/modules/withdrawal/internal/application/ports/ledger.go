package ports

import (
	"context"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger"
)

// LedgerDebiter debits merchant available balance for a withdrawal.
type LedgerDebiter interface {
	PostDebitJournal(ctx context.Context, in ledger.PostDebitJournalInput) (ledger.PostDebitJournalResult, error)
}
