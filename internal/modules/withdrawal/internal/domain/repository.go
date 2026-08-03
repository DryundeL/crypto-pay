package domain

import "context"

// Repository is the write-side persistence port for Withdrawal aggregate.
type Repository interface {
	Save(ctx context.Context, withdrawal *Withdrawal) error
	FindByID(ctx context.Context, id WithdrawalID) (*Withdrawal, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*Withdrawal, error)
}
