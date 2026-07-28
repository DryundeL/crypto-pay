package domain

import "context"

// Repository is the write-side persistence port for Payment aggregate.
type Repository interface {
	Save(ctx context.Context, payment *Payment) error
	FindByID(ctx context.Context, id PaymentID) (*Payment, error)
	FindByNetworkTxHash(ctx context.Context, network, txHash string) (*Payment, error)
}
