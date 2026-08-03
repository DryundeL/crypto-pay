package read

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/domain"
)

type WithdrawalQueries struct {
	db *gorm.DB
}

func NewWithdrawalQueries(db *gorm.DB) *WithdrawalQueries {
	return &WithdrawalQueries{db: db}
}

type withdrawalDTORow struct {
	ID             string    `gorm:"column:id"`
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

func (withdrawalDTORow) TableName() string { return "withdrawal_withdrawals" }

func (q *WithdrawalQueries) GetWithdrawal(ctx context.Context, withdrawalID, merchantID string) (query.WithdrawalDTO, error) {
	var row withdrawalDTORow
	err := q.db.WithContext(ctx).
		Where("id = ? AND merchant_id = ?", withdrawalID, merchantID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.WithdrawalDTO{}, domain.ErrWithdrawalNotFound
		}
		return query.WithdrawalDTO{}, fmt.Errorf("query withdrawal: %w", err)
	}
	return mapDTO(row), nil
}

func (q *WithdrawalQueries) ListWithdrawals(ctx context.Context, merchantID, status string, limit int) ([]query.WithdrawalDTO, error) {
	db := q.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit)
	if status != "" {
		db = db.Where("status = ?", status)
	}

	var rows []withdrawalDTORow
	if err := db.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list withdrawals: %w", err)
	}

	out := make([]query.WithdrawalDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapDTO(row))
	}
	return out, nil
}

func mapDTO(row withdrawalDTORow) query.WithdrawalDTO {
	dto := query.WithdrawalDTO{
		ID:             row.ID,
		MerchantID:     row.MerchantID,
		IdempotencyKey: row.IdempotencyKey,
		Amount:         row.Amount,
		Currency:       row.Currency,
		Network:        row.Network,
		ToAddress:      row.ToAddress,
		Status:         row.Status,
		CreatedAt:      row.CreatedAt.UTC(),
		UpdatedAt:      row.UpdatedAt.UTC(),
	}
	if row.TxHash != nil {
		dto.TxHash = *row.TxHash
	}
	if row.JournalID != nil {
		dto.JournalID = *row.JournalID
	}
	return dto
}
