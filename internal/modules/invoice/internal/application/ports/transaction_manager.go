package ports

import "context"

// TransactionManager abstracts unit-of-work boundaries for write use-cases.
// Implemented by platform/transaction. Aggregate Save + outbox Append must share one TX.
type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
