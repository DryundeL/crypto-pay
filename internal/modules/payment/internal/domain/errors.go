package domain

import "errors"

var (
	ErrInvalidPayment             = errors.New("invalid payment")
	ErrPaymentNotFound            = errors.New("payment not found")
	ErrInvalidTransition          = errors.New("invalid payment status transition")
	ErrInsufficientConfirmations  = errors.New("insufficient confirmations")
)
