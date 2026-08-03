package domain

import "context"

// AccountRepository persists ledger accounts (write side).
type AccountRepository interface {
	FindByKey(ctx context.Context, ownerType OwnerType, ownerID string, kind AccountKind, currency string) (*Account, error)
	Save(ctx context.Context, account *Account) error
}

// JournalRepository persists journals and entries (write side).
type JournalRepository interface {
	FindByIdempotencyKey(ctx context.Context, key string) (*Journal, error)
	Save(ctx context.Context, journal *Journal) error
}
