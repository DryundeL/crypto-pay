package command

import "time"

type withdrawalRequestedEvent struct {
	WithdrawalID string    `json:"withdrawal_id"`
	MerchantID   string    `json:"merchant_id"`
	Amount       string    `json:"amount"`
	Currency     string    `json:"currency"`
	Network      string    `json:"network"`
	OccurredAt   time.Time `json:"occurred_at"`
}

func (withdrawalRequestedEvent) EventName() string { return "withdrawal.requested" }

type withdrawalCompletedEvent struct {
	WithdrawalID string    `json:"withdrawal_id"`
	MerchantID   string    `json:"merchant_id"`
	TxHash       string    `json:"tx_hash"`
	OccurredAt   time.Time `json:"occurred_at"`
}

func (withdrawalCompletedEvent) EventName() string { return "withdrawal.completed" }
