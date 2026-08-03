package read

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/domain"
)

type LedgerQueries struct {
	db *gorm.DB
}

func NewLedgerQueries(db *gorm.DB) *LedgerQueries {
	return &LedgerQueries{db: db}
}

type balanceRow struct {
	ID       string `gorm:"column:id"`
	Currency string `gorm:"column:currency"`
	Kind     string `gorm:"column:kind"`
	Balance  string `gorm:"column:balance"`
}

func (balanceRow) TableName() string { return "ledger_accounts" }

type journalDTORow struct {
	ID            string    `gorm:"column:id"`
	MerchantID    string    `gorm:"column:merchant_id"`
	ReferenceType string    `gorm:"column:reference_type"`
	ReferenceID   string    `gorm:"column:reference_id"`
	Amount        string    `gorm:"column:amount"`
	Currency      string    `gorm:"column:currency"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (journalDTORow) TableName() string { return "ledger_journals" }

func (q *LedgerQueries) ListBalances(ctx context.Context, merchantID string) ([]query.BalanceDTO, error) {
	var rows []balanceRow
	err := q.db.WithContext(ctx).
		Select("id", "currency", "kind", "balance").
		Where("owner_type = ? AND owner_id = ?", domain.OwnerMerchant.String(), merchantID).
		Order("currency ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list balances: %w", err)
	}

	out := make([]query.BalanceDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, query.BalanceDTO{
			AccountID: row.ID,
			Currency:  row.Currency,
			Kind:      row.Kind,
			Balance:   row.Balance,
		})
	}
	return out, nil
}

func (q *LedgerQueries) ListJournals(ctx context.Context, merchantID, currency string, limit int) ([]query.JournalDTO, error) {
	db := q.db.WithContext(ctx).
		Select("id", "merchant_id", "reference_type", "reference_id", "amount", "currency", "created_at").
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit)
	if currency != "" {
		db = db.Where("currency = ?", currency)
	}

	var rows []journalDTORow
	if err := db.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list journals: %w", err)
	}

	out := make([]query.JournalDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, query.JournalDTO{
			ID:            row.ID,
			MerchantID:    row.MerchantID,
			ReferenceType: row.ReferenceType,
			ReferenceID:   row.ReferenceID,
			Amount:        row.Amount,
			Currency:      row.Currency,
			CreatedAt:     row.CreatedAt.UTC(),
		})
	}
	return out, nil
}
