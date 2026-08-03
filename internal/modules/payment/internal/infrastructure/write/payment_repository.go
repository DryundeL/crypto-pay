package write

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

type paymentRow struct {
	ID            string    `gorm:"column:id;primaryKey"`
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

func (paymentRow) TableName() string { return "payments" }

func (r *PaymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return err
	}

	row := paymentRow{
		ID:            payment.ID().String(),
		InvoiceID:     payment.InvoiceID(),
		MerchantID:    payment.MerchantID(),
		Status:        payment.Status().String(),
		Network:       payment.Network(),
		TxHash:        payment.TxHash(),
		ToAddress:     payment.ToAddress(),
		Amount:        payment.Amount(),
		Currency:      payment.Currency(),
		Confirmations: payment.Confirmations(),
		CreatedAt:     payment.CreatedAt().UTC(),
		UpdatedAt:     payment.UpdatedAt().UTC(),
	}

	if err := db.Save(&row).Error; err != nil {
		return fmt.Errorf("upsert payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) FindByID(ctx context.Context, id domain.PaymentID) (*domain.Payment, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var row paymentRow
	if err := db.Where("id = ?", id.String()).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("load payment: %w", err)
	}
	return mapPayment(row), nil
}

func (r *PaymentRepository) FindByNetworkTxHash(ctx context.Context, network, txHash string) (*domain.Payment, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var row paymentRow
	if err := db.Where("network = ? AND tx_hash = ?", network, txHash).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("load payment by tx: %w", err)
	}
	return mapPayment(row), nil
}

func mapPayment(row paymentRow) *domain.Payment {
	return domain.Restore(
		domain.PaymentID(row.ID),
		row.InvoiceID,
		row.MerchantID,
		row.Network,
		row.TxHash,
		row.ToAddress,
		row.Amount,
		row.Currency,
		row.Confirmations,
		domain.Status(row.Status),
		row.CreatedAt,
		row.UpdatedAt,
	)
}
