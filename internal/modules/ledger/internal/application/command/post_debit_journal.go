package command

type PostDebitJournal struct {
	IdempotencyKey string
	MerchantID     string
	Amount         string
	Currency       string
	ReferenceType  string
	ReferenceID    string
}

type PostDebitJournalResult struct {
	JournalID string
	Created   bool
}
