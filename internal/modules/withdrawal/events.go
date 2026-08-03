package withdrawal

import "time"

const (
	EventWithdrawalRequested = "withdrawal.requested"
	EventWithdrawalCompleted = "withdrawal.completed"
)

type WithdrawalRequested struct {
	WithdrawalID string    `json:"withdrawal_id"`
	MerchantID   string    `json:"merchant_id"`
	Amount       string    `json:"amount"`
	Currency     string    `json:"currency"`
	Network      string    `json:"network"`
	OccurredAt   time.Time `json:"occurred_at"`
}

func (WithdrawalRequested) EventName() string { return EventWithdrawalRequested }

type WithdrawalCompleted struct {
	WithdrawalID string    `json:"withdrawal_id"`
	MerchantID   string    `json:"merchant_id"`
	TxHash       string    `json:"tx_hash"`
	OccurredAt   time.Time `json:"occurred_at"`
}

func (WithdrawalCompleted) EventName() string { return EventWithdrawalCompleted }
