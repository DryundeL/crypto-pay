package write

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

type InvoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

type invoiceRow struct {
	ID         string    `gorm:"column:id;primaryKey"`
	MerchantID string    `gorm:"column:merchant_id"`
	Status     string    `gorm:"column:status"`
	Amount     string    `gorm:"column:amount"`
	Currency   string    `gorm:"column:currency"`
	Network    string    `gorm:"column:network"`
	Address    string    `gorm:"column:address"`
	TxHash     *string   `gorm:"column:tx_hash"`
	ExpiresAt  time.Time `gorm:"column:expires_at"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (invoiceRow) TableName() string { return "invoices" }

func (r *InvoiceRepository) Save(ctx context.Context, invoice *domain.Invoice) error {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return err
	}

	var txHash *string
	if invoice.TxHash() != "" {
		h := invoice.TxHash()
		txHash = &h
	}

	row := invoiceRow{
		ID:         invoice.ID().String(),
		MerchantID: invoice.MerchantID(),
		Status:     invoice.Status().String(),
		Amount:     invoice.Amount().Amount(),
		Currency:   invoice.Amount().Currency(),
		Network:    invoice.Network(),
		Address:    invoice.Address(),
		TxHash:     txHash,
		ExpiresAt:  invoice.ExpiresAt().UTC(),
		CreatedAt:  invoice.CreatedAt().UTC(),
		UpdatedAt:  invoice.UpdatedAt().UTC(),
	}

	if err := db.Save(&row).Error; err != nil {
		return fmt.Errorf("upsert invoice: %w", err)
	}
	return nil
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id domain.InvoiceID) (*domain.Invoice, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var row invoiceRow
	if err := db.Where("id = ?", id.String()).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("load invoice: %w", err)
	}

	money, err := domain.NewMoney(row.Amount, row.Currency)
	if err != nil {
		return nil, fmt.Errorf("restore money: %w", err)
	}

	txHash := ""
	if row.TxHash != nil {
		txHash = *row.TxHash
	}

	return domain.Restore(
		domain.InvoiceID(row.ID),
		row.MerchantID,
		domain.Status(row.Status),
		money,
		row.Network,
		row.Address,
		txHash,
		row.ExpiresAt,
		row.CreatedAt,
		row.UpdatedAt,
	), nil
}

func (r *InvoiceRepository) ListExpiredPendingIDs(ctx context.Context, now time.Time, limit int) ([]domain.InvoiceID, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}

	var ids []string
	err = db.Model(&invoiceRow{}).
		Select("id").
		Where("status = ? AND expires_at < ?", domain.StatusPending.String(), now.UTC()).
		Order("expires_at ASC").
		Limit(limit).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("list expired pending invoices: %w", err)
	}

	out := make([]domain.InvoiceID, 0, len(ids))
	for _, id := range ids {
		out = append(out, domain.InvoiceID(id))
	}
	return out, nil
}
