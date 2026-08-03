package domain

import "fmt"

// EntrySide is debit or credit in double-entry bookkeeping.
type EntrySide string

const (
	SideDebit  EntrySide = "debit"
	SideCredit EntrySide = "credit"
)

func (s EntrySide) String() string { return string(s) }

func (s EntrySide) Valid() bool {
	return s == SideDebit || s == SideCredit
}

func ParseEntrySide(v string) (EntrySide, error) {
	s := EntrySide(v)
	if !s.Valid() {
		return "", fmt.Errorf("%w: invalid entry side %q", ErrInvalidJournal, v)
	}
	return s, nil
}
