package ports

import "context"

// TransactionManager runs aggregate Save + outbox Append in one DB transaction.
type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
