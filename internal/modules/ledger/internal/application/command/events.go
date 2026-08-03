package command

import "time"

// Integration event payloads for outbox (mirror public ledger.events contracts).

type entryPostedEvent struct {
	JournalID     string    `json:"journal_id"`
	MerchantID    string    `json:"merchant_id"`
	Amount        string    `json:"amount"`
	Currency      string    `json:"currency"`
	ReferenceType string    `json:"reference_type"`
	ReferenceID   string    `json:"reference_id"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (entryPostedEvent) EventName() string { return "ledger.entry_posted" }
