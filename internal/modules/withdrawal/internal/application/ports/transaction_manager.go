package ports

import "context"

// TransactionManager abstracts unit-of-work boundaries for write use-cases.
type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
