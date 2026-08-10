package domain

import "errors"

var (
	ErrInvalidBlockchain           = errors.New("invalid blockchain input")
	ErrAddressNotFound             = errors.New("blockchain address not found")
	ErrTxNotFound                  = errors.New("blockchain transaction not found")
	ErrInvalidTransition           = errors.New("invalid blockchain transaction status transition")
	ErrInsufficientConfirmations   = errors.New("insufficient confirmations")
)
