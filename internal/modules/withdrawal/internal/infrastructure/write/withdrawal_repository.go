package write

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

type WithdrawalRepository struct {
	db *gorm.DB
}

func NewWithdrawalRepository(db *gorm.DB) *WithdrawalRepository {
	return &WithdrawalRepository{db: db}
}

type withdrawalRow struct {
	ID             string    `gorm:"column:id;primaryKey"`
	MerchantID     string    `gorm:"column:merchant_id"`
	IdempotencyKey string    `gorm:"column:idempotency_key"`
	Amount         string    `gorm:"column:amount"`
	Currency       string    `gorm:"column:currency"`
	Network        string    `gorm:"column:network"`
	ToAddress      string    `gorm:"column:to_address"`
	Status         string    `gorm:"column:status"`
	TxHash         *string   `gorm:"column:tx_hash"`
	JournalID      *string   `gorm:"column:journal_id"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (withdrawalRow) TableName() string { return "withdrawal_withdrawals" }

func (r *WithdrawalRepository) Save(ctx context.Context, w *domain.Withdrawal) error {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return err
	}

	row := withdrawalRow{
		ID:             w.ID().String(),
		MerchantID:     w.MerchantID(),
		IdempotencyKey: w.IdempotencyKey(),
		Amount:         w.Amount(),
		Currency:       w.Currency(),
		Network:        w.Network(),
		ToAddress:      w.ToAddress(),
		Status:         w.Status().String(),
		CreatedAt:      w.CreatedAt().UTC(),
		UpdatedAt:      w.UpdatedAt().UTC(),
	}
	if w.TxHash() != "" {
		h := w.TxHash()
		row.TxHash = &h
	}
	if w.JournalID() != "" {
		j := w.JournalID()
		row.JournalID = &j
	}

	if err := db.Save(&row).Error; err != nil {
		return fmt.Errorf("upsert withdrawal: %w", err)
	}
	return nil
}

func (r *WithdrawalRepository) FindByID(ctx context.Context, id domain.WithdrawalID) (*domain.Withdrawal, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var row withdrawalRow
	if err := db.Where("id = ?", id.String()).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrWithdrawalNotFound
		}
		return nil, fmt.Errorf("find withdrawal: %w", err)
	}
	return mapWithdrawal(row), nil
}

func (r *WithdrawalRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Withdrawal, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var row withdrawalRow
	if err := db.Where("idempotency_key = ?", key).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrWithdrawalNotFound
		}
		return nil, fmt.Errorf("find withdrawal by idempotency key: %w", err)
	}
	return mapWithdrawal(row), nil
}

func mapWithdrawal(row withdrawalRow) *domain.Withdrawal {
	txHash := ""
	if row.TxHash != nil {
		txHash = *row.TxHash
	}
	journalID := ""
	if row.JournalID != nil {
		journalID = *row.JournalID
	}
	return domain.Restore(
		domain.WithdrawalID(row.ID),
		row.MerchantID,
		row.IdempotencyKey,
		row.Amount,
		row.Currency,
		row.Network,
		row.ToAddress,
		domain.Status(row.Status),
		txHash,
		journalID,
		row.CreatedAt,
		row.UpdatedAt,
	)
}
