package withdrawal

import "errors"

// Public contracts of the Withdrawal bounded context.

type Status string

const (
	StatusRequested Status = "requested"
	StatusApproved  Status = "approved"
	StatusBroadcast Status = "broadcast"
	StatusCompleted Status = "completed"
	StatusRejected  Status = "rejected"
)

var (
	ErrNotFound            = errors.New("withdrawal not found")
	ErrInvalidInput        = errors.New("invalid withdrawal input")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidTransition   = errors.New("invalid withdrawal status transition")
)
