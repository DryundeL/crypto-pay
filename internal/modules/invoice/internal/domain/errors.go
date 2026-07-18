package domain

import "errors"

var (
	ErrInvalidMoney     = errors.New("invalid money")
	ErrInvoiceNotFound  = errors.New("invoice not found")
	ErrInvalidTransition = errors.New("invalid invoice status transition")
)
