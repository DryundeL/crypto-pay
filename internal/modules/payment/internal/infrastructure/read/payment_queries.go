package read

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/domain"
)

type PaymentQueries struct {
	db *gorm.DB
}

func NewPaymentQueries(db *gorm.DB) *PaymentQueries {
	return &PaymentQueries{db: db}
}

type paymentDTORow struct {
	ID            string    `gorm:"column:id"`
	InvoiceID     string    `gorm:"column:invoice_id"`
	MerchantID    string    `gorm:"column:merchant_id"`
	Status        string    `gorm:"column:status"`
	Network       string    `gorm:"column:network"`
	TxHash        string    `gorm:"column:tx_hash"`
	ToAddress     string    `gorm:"column:to_address"`
	Amount        string    `gorm:"column:amount"`
	Currency      string    `gorm:"column:currency"`
	Confirmations int       `gorm:"column:confirmations"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (paymentDTORow) TableName() string { return "payments" }

func (q *PaymentQueries) GetPayment(ctx context.Context, paymentID, merchantID string) (query.PaymentDTO, error) {
	var row paymentDTORow
	err := q.db.WithContext(ctx).
		Select("id", "invoice_id", "merchant_id", "status", "network", "tx_hash", "to_address", "amount", "currency", "confirmations", "created_at", "updated_at").
		Where("id = ? AND merchant_id = ?", paymentID, merchantID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.PaymentDTO{}, domain.ErrPaymentNotFound
		}
		return query.PaymentDTO{}, fmt.Errorf("query payment: %w", err)
	}

	return query.PaymentDTO{
		ID:            row.ID,
		InvoiceID:     row.InvoiceID,
		MerchantID:    row.MerchantID,
		Status:        row.Status,
		Network:       row.Network,
		TxHash:        row.TxHash,
		ToAddress:     row.ToAddress,
		Amount:        row.Amount,
		Currency:      row.Currency,
		Confirmations: row.Confirmations,
		CreatedAt:     row.CreatedAt.UTC(),
		UpdatedAt:     row.UpdatedAt.UTC(),
	}, nil
}
