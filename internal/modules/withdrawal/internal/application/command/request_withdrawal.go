package command

import "time"

type RequestWithdrawal struct {
	MerchantID     string
	IdempotencyKey string
	Amount         string
	Currency       string
	Network        string
	ToAddress      string
}

type RequestWithdrawalResult struct {
	ID             string
	MerchantID     string
	IdempotencyKey string
	Amount         string
	Currency       string
	Network        string
	ToAddress      string
	Status         string
	JournalID      string
	Created        bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
