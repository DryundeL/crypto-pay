package domain

import (
	"context"
	"time"
)

// Repository is the write-side persistence port for Delivery aggregate.
type Repository interface {
	Save(ctx context.Context, delivery *Delivery) error
	FindByID(ctx context.Context, id DeliveryID) (*Delivery, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*Delivery, error)
	FindDue(ctx context.Context, now time.Time, limit int) ([]*Delivery, error)
}
