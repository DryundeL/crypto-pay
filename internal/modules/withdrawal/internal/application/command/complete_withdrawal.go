package command

import "time"

type CompleteWithdrawal struct {
	WithdrawalID string
	TxHash       string
}

type CompleteWithdrawalResult struct {
	ID         string
	MerchantID string
	Status     string
	TxHash     string
	Amount     string
	Currency   string
	Network    string
	ToAddress  string
	JournalID  string
	UpdatedAt  time.Time
}
