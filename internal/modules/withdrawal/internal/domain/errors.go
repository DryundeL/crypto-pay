package domain

import "errors"

var (
	ErrInvalidWithdrawal  = errors.New("invalid withdrawal")
	ErrWithdrawalNotFound = errors.New("withdrawal not found")
	ErrInvalidTransition  = errors.New("invalid withdrawal status transition")
)
