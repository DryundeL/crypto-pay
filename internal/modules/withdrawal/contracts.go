package withdrawal

import (
	"context"
	"errors"
	"time"
)

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

// WithdrawalCompletedNotification is a composition-root side-effect payload after Complete.
type WithdrawalCompletedNotification struct {
	WithdrawalID string
	MerchantID   string
	TxHash       string
	Amount       string
	Currency     string
	Network      string
	OccurredAt   time.Time
}

// CompletedNotifier is invoked by Module.Complete after a successful transition.
type CompletedNotifier interface {
	NotifyWithdrawalCompleted(ctx context.Context, in WithdrawalCompletedNotification) error
}
