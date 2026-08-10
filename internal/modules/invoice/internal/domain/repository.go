package domain

import (
	"context"
	"time"
)

// Repository is the write-side persistence port for Invoice aggregate.
type Repository interface {
	Save(ctx context.Context, invoice *Invoice) error
	FindByID(ctx context.Context, id InvoiceID) (*Invoice, error)
	ListExpiredPendingIDs(ctx context.Context, now time.Time, limit int) ([]InvoiceID, error)
}
