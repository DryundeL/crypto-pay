package ledger

// Public contracts of the Ledger bounded context.
type EntrySide string

const (
	EntryDebit  EntrySide = "debit"
	EntryCredit EntrySide = "credit"
)
