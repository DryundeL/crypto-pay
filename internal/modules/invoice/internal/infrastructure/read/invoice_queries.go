package read

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/domain"
)

type InvoiceQueries struct {
	db *gorm.DB
}

func NewInvoiceQueries(db *gorm.DB) *InvoiceQueries {
	return &InvoiceQueries{db: db}
}

type invoiceDTORow struct {
	ID         string    `gorm:"column:id"`
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

func (invoiceDTORow) TableName() string { return "invoice_invoices" }

func (q *InvoiceQueries) GetInvoice(ctx context.Context, invoiceID, merchantID string) (query.InvoiceDTO, error) {
	var row invoiceDTORow
	err := q.db.WithContext(ctx).
		Select("id", "merchant_id", "status", "amount", "currency", "network", "address", "tx_hash", "expires_at", "created_at", "updated_at").
		Where("id = ? AND merchant_id = ?", invoiceID, merchantID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.InvoiceDTO{}, domain.ErrInvoiceNotFound
		}
		return query.InvoiceDTO{}, fmt.Errorf("query invoice: %w", err)
	}
	return mapInvoiceDTO(row), nil
}

func (q *InvoiceQueries) FindPendingByAddress(ctx context.Context, network, address string) (query.InvoiceDTO, error) {
	var row invoiceDTORow
	err := q.db.WithContext(ctx).
		Select("id", "merchant_id", "status", "amount", "currency", "network", "address", "tx_hash", "expires_at", "created_at", "updated_at").
		Where("network = ? AND address = ? AND status = ?", network, address, domain.StatusPending.String()).
		Order("created_at DESC").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.InvoiceDTO{}, domain.ErrInvoiceNotFound
		}
		return query.InvoiceDTO{}, fmt.Errorf("query invoice by address: %w", err)
	}
	return mapInvoiceDTO(row), nil
}

func mapInvoiceDTO(row invoiceDTORow) query.InvoiceDTO {
	txHash := ""
	if row.TxHash != nil {
		txHash = *row.TxHash
	}
	return query.InvoiceDTO{
		ID:         row.ID,
		MerchantID: row.MerchantID,
		Status:     row.Status,
		Amount:     row.Amount,
		Currency:   row.Currency,
		Network:    row.Network,
		Address:    row.Address,
		TxHash:     txHash,
		ExpiresAt:  row.ExpiresAt.UTC(),
		CreatedAt:  row.CreatedAt.UTC(),
		UpdatedAt:  row.UpdatedAt.UTC(),
	}
}
