package domain

import "errors"

var (
	ErrInvalidMoney        = errors.New("invalid money")
	ErrInvalidAccount      = errors.New("invalid account")
	ErrInvalidJournal      = errors.New("invalid journal")
	ErrAccountNotFound     = errors.New("account not found")
	ErrJournalNotFound     = errors.New("journal not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
)
