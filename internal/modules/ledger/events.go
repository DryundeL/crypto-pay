package ledger

import "time"

const EventEntryPosted = "ledger.entry_posted"

type EntryPosted struct {
	JournalID  string    `json:"journal_id"`
	MerchantID string    `json:"merchant_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (EntryPosted) EventName() string { return EventEntryPosted }
