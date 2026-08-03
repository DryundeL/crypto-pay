package ledger

import "errors"

// Public contracts of the Ledger bounded context.
// Other modules depend only on these types, never on ledger/internal.

type EntrySide string

const (
	EntryDebit  EntrySide = "debit"
	EntryCredit EntrySide = "credit"
)

type AccountKind string

const (
	AccountKindClearing  AccountKind = "clearing"
	AccountKindAvailable AccountKind = "available"
)

// BalanceView is a cross-module read projection of merchant available balance.
type BalanceView struct {
	AccountID string
	Currency  string
	Kind      string
	Balance   string
}

// PostJournalInput is the sync facade command for posting a credit journal.
type PostJournalInput struct {
	IdempotencyKey string
	MerchantID     string
	Amount         string
	Currency       string
	ReferenceType  string
	ReferenceID    string
}

// PostJournalResult is returned by Module.PostJournal.
type PostJournalResult struct {
	JournalID string
	Created   bool
}

var (
	ErrNotFound            = errors.New("ledger resource not found")
	ErrInvalidInput        = errors.New("invalid ledger input")
	ErrInsufficientBalance = errors.New("insufficient balance")
)
