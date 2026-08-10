package command

import "time"

type MarkPaid struct {
	InvoiceID string
	TxHash    string
}

type MarkPaidResult struct {
	InvoiceID  string
	MerchantID string
	Amount     string
	Currency   string
	TxHash     string
	OccurredAt time.Time
}
