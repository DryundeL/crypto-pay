package transaction

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Manager runs write use-cases inside a single DB transaction
// (aggregate persistence + outbox append).
type Manager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type GormManager struct {
	db *gorm.DB
}

func NewGormManager(db *gorm.DB) *GormManager {
	return &GormManager{db: db}
}

func (m *GormManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := withTx(ctx, tx)
		if err := fn(txCtx); err != nil {
			return err
		}
		return nil
	})
}

type txKey struct{}

func withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext returns the transactional *gorm.DB if present.
func TxFromContext(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	return tx, ok
}

// DB returns tx from context or the fallback handle.
func DB(ctx context.Context, fallback *gorm.DB) (*gorm.DB, error) {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		return tx, nil
	}
	if fallback == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	return fallback.WithContext(ctx), nil
}
