package domain

import "context"

// Repository is the write-side persistence port for Merchant aggregate.
type Repository interface {
	Save(ctx context.Context, merchant *Merchant) error
	FindByID(ctx context.Context, id MerchantID) (*Merchant, error)
}
