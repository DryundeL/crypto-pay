package query

type ListJournals struct {
	MerchantID string
	Currency   string // optional filter
	Limit      int
}
