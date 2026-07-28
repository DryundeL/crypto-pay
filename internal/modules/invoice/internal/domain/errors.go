package domain

import "errors"

var (
	ErrInvalidMoney      = errors.New("invalid money")
	ErrInvalidInvoice    = errors.New("invalid invoice")
	ErrInvoiceNotFound   = errors.New("invoice not found")
	ErrInvalidTransition = errors.New("invalid invoice status transition")
	ErrInvoiceNotExpired = errors.New("invoice has not expired yet")
)
