package domain

import (
	"fmt"
	"strings"
	"time"
)

// Entry is a single debit/credit leg of a journal.
type Entry struct {
	id        EntryID
	accountID AccountID
	side      EntrySide
	amount    Money
}

func (e Entry) ID() EntryID          { return e.id }
func (e Entry) AccountID() AccountID { return e.accountID }
func (e Entry) Side() EntrySide      { return e.side }
func (e Entry) Amount() Money        { return e.amount }

// Journal is the aggregate root: a balanced double-entry posting.
type Journal struct {
	id             JournalID
	idempotencyKey string
	merchantID     string
	referenceType  string
	referenceID    string
	entries        []Entry
	createdAt      time.Time
}

type PostParams struct {
	ID                JournalID
	IdempotencyKey    string
	MerchantID        string
	Amount            Money
	ReferenceType     string
	ReferenceID       string
	ClearingAccountID AccountID
	MerchantAccountID AccountID
	ClearingEntryID   EntryID
	MerchantEntryID   EntryID
	Now               time.Time
}

// Post creates a balanced 2-leg credit journal:
// debit system clearing, credit merchant available.
func Post(p PostParams) (*Journal, error) {
	if p.ID.IsZero() {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidJournal)
	}
	key := strings.TrimSpace(p.IdempotencyKey)
	if key == "" {
		return nil, fmt.Errorf("%w: idempotency_key is required", ErrInvalidJournal)
	}
	if len(key) > 128 {
		return nil, fmt.Errorf("%w: idempotency_key too long", ErrInvalidJournal)
	}
	merchantID := strings.TrimSpace(p.MerchantID)
	if merchantID == "" {
		return nil, fmt.Errorf("%w: merchant_id is required", ErrInvalidJournal)
	}
	refType := strings.TrimSpace(p.ReferenceType)
	if refType == "" {
		return nil, fmt.Errorf("%w: reference_type is required", ErrInvalidJournal)
	}
	refID := strings.TrimSpace(p.ReferenceID)
	if refID == "" {
		return nil, fmt.Errorf("%w: reference_id is required", ErrInvalidJournal)
	}
	if p.Amount.IsZero() {
		return nil, fmt.Errorf("%w: amount must be positive", ErrInvalidJournal)
	}
	if p.ClearingAccountID.IsZero() || p.MerchantAccountID.IsZero() {
		return nil, fmt.Errorf("%w: account ids are required", ErrInvalidJournal)
	}
	if p.ClearingEntryID.IsZero() || p.MerchantEntryID.IsZero() {
		return nil, fmt.Errorf("%w: entry ids are required", ErrInvalidJournal)
	}
	if p.ClearingAccountID == p.MerchantAccountID {
		return nil, fmt.Errorf("%w: accounts must differ", ErrInvalidJournal)
	}

	entries := []Entry{
		{
			id:        p.ClearingEntryID,
			accountID: p.ClearingAccountID,
			side:      SideDebit,
			amount:    p.Amount,
		},
		{
			id:        p.MerchantEntryID,
			accountID: p.MerchantAccountID,
			side:      SideCredit,
			amount:    p.Amount,
		},
	}
	if err := assertBalanced(entries); err != nil {
		return nil, err
	}

	return &Journal{
		id:             p.ID,
		idempotencyKey: key,
		merchantID:     merchantID,
		referenceType:  refType,
		referenceID:    refID,
		entries:        entries,
		createdAt:      p.Now.UTC(),
	}, nil
}

func RestoreJournal(
	id JournalID,
	idempotencyKey, merchantID, referenceType, referenceID string,
	entries []Entry,
	createdAt time.Time,
) *Journal {
	return &Journal{
		id:             id,
		idempotencyKey: idempotencyKey,
		merchantID:     merchantID,
		referenceType:  referenceType,
		referenceID:    referenceID,
		entries:        entries,
		createdAt:      createdAt,
	}
}

func RestoreEntry(id EntryID, accountID AccountID, side EntrySide, amount Money) Entry {
	return Entry{id: id, accountID: accountID, side: side, amount: amount}
}

func (j *Journal) ID() JournalID          { return j.id }
func (j *Journal) IdempotencyKey() string { return j.idempotencyKey }
func (j *Journal) MerchantID() string     { return j.merchantID }
func (j *Journal) ReferenceType() string  { return j.referenceType }
func (j *Journal) ReferenceID() string    { return j.referenceID }
func (j *Journal) Entries() []Entry       { return append([]Entry(nil), j.entries...) }
func (j *Journal) CreatedAt() time.Time   { return j.createdAt }
func (j *Journal) Amount() Money          { return j.entries[0].amount }
func (j *Journal) Currency() string       { return j.entries[0].amount.Currency() }

func assertBalanced(entries []Entry) error {
	if len(entries) != 2 {
		return fmt.Errorf("%w: journal must have exactly 2 entries", ErrInvalidJournal)
	}
	a, b := entries[0], entries[1]
	if a.side == b.side {
		return fmt.Errorf("%w: entries must have opposite sides", ErrInvalidJournal)
	}
	if !a.side.Valid() || !b.side.Valid() {
		return fmt.Errorf("%w: invalid entry side", ErrInvalidJournal)
	}
	if !a.amount.Equal(b.amount) {
		return fmt.Errorf("%w: entry amounts must match", ErrInvalidJournal)
	}
	return nil
}
