package write

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

type txRow struct {
	ID            string    `gorm:"column:id;primaryKey"`
	Network       string    `gorm:"column:network"`
	TxHash        string    `gorm:"column:tx_hash"`
	ToAddress     string    `gorm:"column:to_address"`
	Amount        string    `gorm:"column:amount"`
	Currency      string    `gorm:"column:currency"`
	Confirmations int       `gorm:"column:confirmations"`
	Status        string    `gorm:"column:status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (txRow) TableName() string { return "blockchain_transactions" }

func (r *TransactionRepository) Save(ctx context.Context, tx *domain.WatchedTransaction) error {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return err
	}
	row := txRow{
		ID:            tx.ID(),
		Network:       tx.Network(),
		TxHash:        tx.TxHash(),
		ToAddress:     tx.ToAddress(),
		Amount:        tx.Amount(),
		Currency:      tx.Currency(),
		Confirmations: tx.Confirmations(),
		Status:        tx.Status().String(),
		CreatedAt:     tx.CreatedAt().UTC(),
		UpdatedAt:     tx.UpdatedAt().UTC(),
	}
	if err := db.Save(&row).Error; err != nil {
		return fmt.Errorf("upsert blockchain tx: %w", err)
	}
	return nil
}

func (r *TransactionRepository) FindByNetworkTxHash(ctx context.Context, network, txHash string) (*domain.WatchedTransaction, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var row txRow
	if err := db.Where("network = ? AND tx_hash = ?", network, txHash).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTxNotFound
		}
		return nil, fmt.Errorf("load blockchain tx: %w", err)
	}
	return domain.RestoreWatchedTransaction(
		row.ID, row.Network, row.TxHash, row.ToAddress, row.Amount, row.Currency,
		row.Confirmations, domain.TxStatus(row.Status), row.CreatedAt, row.UpdatedAt,
	), nil
}
