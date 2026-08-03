package command

type PostJournal struct {
	IdempotencyKey string
	MerchantID     string
	Amount         string
	Currency       string
	ReferenceType  string
	ReferenceID    string
}

type PostJournalResult struct {
	JournalID string
	Created   bool
}
