package command

import "time"

type transactionObservedEvent struct {
	Network       string    `json:"network"`
	TxHash        string    `json:"tx_hash"`
	ToAddress     string    `json:"to_address"`
	Amount        string    `json:"amount"`
	Currency      string    `json:"currency"`
	Confirmations int       `json:"confirmations"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (transactionObservedEvent) EventName() string { return "blockchain.transaction_observed" }

type transactionConfirmedEvent struct {
	Network    string    `json:"network"`
	TxHash     string    `json:"tx_hash"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (transactionConfirmedEvent) EventName() string { return "blockchain.transaction_confirmed" }
