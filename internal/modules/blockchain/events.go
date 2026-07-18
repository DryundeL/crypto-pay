package blockchain

import "time"

const (
	EventTransactionObserved  = "blockchain.transaction_observed"
	EventTransactionConfirmed = "blockchain.transaction_confirmed"
)

type TransactionObserved struct {
	Network       string    `json:"network"`
	TxHash        string    `json:"tx_hash"`
	ToAddress     string    `json:"to_address"`
	Amount        string    `json:"amount"`
	Currency      string    `json:"currency"`
	Confirmations int       `json:"confirmations"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (TransactionObserved) EventName() string { return EventTransactionObserved }

type TransactionConfirmed struct {
	Network    string    `json:"network"`
	TxHash     string    `json:"tx_hash"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (TransactionConfirmed) EventName() string { return EventTransactionConfirmed }
