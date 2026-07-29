package ports

import (
	"context"

	"github.com/DryundeL/crypto-pay/internal/modules/payment"
)

// PaymentNotifier sync-notifies the Payment BC after chain observations.
type PaymentNotifier interface {
	RecordObserved(ctx context.Context, in payment.RecordObservedInput) (payment.PaymentView, error)
	Confirm(ctx context.Context, in payment.ConfirmInput) (payment.PaymentView, error)
}
