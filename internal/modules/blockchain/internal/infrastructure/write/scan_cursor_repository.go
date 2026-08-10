package write

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// DepositLookup joins blockchain_addresses with invoices for scanner matching.
type DepositLookup struct {
	db *gorm.DB
}

func NewDepositLookup(db *gorm.DB) *DepositLookup {
	return &DepositLookup{db: db}
}

type depositRow struct {
	Network    string `gorm:"column:network"`
	Address    string `gorm:"column:address"`
	InvoiceID  string `gorm:"column:invoice_id"`
	MerchantID string `gorm:"column:merchant_id"`
	Currency   string `gorm:"column:currency"`
}

func (r *DepositLookup) LookupByNetworkAddress(ctx context.Context, network, address string) (domain.DepositRef, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return domain.DepositRef{}, err
	}
	var row depositRow
	err = db.WithContext(ctx).Raw(`
		SELECT a.network, a.address, a.invoice_id, i.merchant_id, a.currency
		FROM blockchain_addresses a
		INNER JOIN invoices i ON i.id = a.invoice_id
		WHERE a.network = ? AND lower(a.address) = lower(?)
		LIMIT 1
	`, network, strings.TrimSpace(address)).Scan(&row).Error
	if err != nil {
		return domain.DepositRef{}, fmt.Errorf("lookup deposit: %w", err)
	}
	if row.InvoiceID == "" || row.MerchantID == "" {
		return domain.DepositRef{}, domain.ErrAddressNotFound
	}
	return domain.DepositRef{
		Network:    row.Network,
		Address:    row.Address,
		InvoiceID:  row.InvoiceID,
		MerchantID: row.MerchantID,
		Currency:   row.Currency,
	}, nil
}

// ScanCursorRepository persists scanner tip checkpoints.
type ScanCursorRepository struct {
	db *gorm.DB
}

func NewScanCursorRepository(db *gorm.DB) *ScanCursorRepository {
	return &ScanCursorRepository{db: db}
}

type scanCursorRow struct {
	Network     string    `gorm:"column:network;primaryKey"`
	BlockNumber int64     `gorm:"column:block_number"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (scanCursorRow) TableName() string { return "blockchain_scan_cursors" }

func (r *ScanCursorRepository) Get(ctx context.Context, network string) (uint64, bool, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return 0, false, err
	}
	var row scanCursorRow
	if err := db.Where("network = ?", network).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("load scan cursor: %w", err)
	}
	if row.BlockNumber < 0 {
		return 0, false, fmt.Errorf("scan cursor block_number is negative")
	}
	return uint64(row.BlockNumber), true, nil
}

func (r *ScanCursorRepository) Upsert(ctx context.Context, network string, blockNumber uint64) error {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return err
	}
	row := scanCursorRow{
		Network:     network,
		BlockNumber: int64(blockNumber),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "network"}},
		DoUpdates: clause.AssignmentColumns([]string{"block_number", "updated_at"}),
	}).Create(&row).Error; err != nil {
		return fmt.Errorf("upsert scan cursor: %w", err)
	}
	return nil
}
