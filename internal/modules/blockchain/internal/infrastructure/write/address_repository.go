package write

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

type AddressRepository struct {
	db *gorm.DB
}

func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{db: db}
}

type addressRow struct {
	ID             string    `gorm:"column:id;primaryKey"`
	Network        string    `gorm:"column:network"`
	Address        string    `gorm:"column:address"`
	DerivationPath string    `gorm:"column:derivation_path"`
	InvoiceID      string    `gorm:"column:invoice_id"`
	Currency       string    `gorm:"column:currency"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (addressRow) TableName() string { return "blockchain_addresses" }

type counterRow struct {
	Network   string `gorm:"column:network;primaryKey"`
	NextIndex int64  `gorm:"column:next_index"`
}

func (counterRow) TableName() string { return "blockchain_network_counters" }

func (r *AddressRepository) FindByNetworkInvoice(ctx context.Context, network, invoiceID string) (*domain.AllocatedAddress, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var row addressRow
	if err := db.Where("network = ? AND invoice_id = ?", network, invoiceID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAddressNotFound
		}
		return nil, fmt.Errorf("load address: %w", err)
	}
	return domain.RestoreAllocatedAddress(
		row.ID, row.Network, row.Address, row.DerivationPath, row.InvoiceID, row.Currency, row.CreatedAt,
	), nil
}

func (r *AddressRepository) FindByNetworkAddress(ctx context.Context, network, address string) (*domain.AllocatedAddress, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var row addressRow
	if err := db.Where("network = ? AND lower(address) = lower(?)", network, address).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAddressNotFound
		}
		return nil, fmt.Errorf("load address by address: %w", err)
	}
	return domain.RestoreAllocatedAddress(
		row.ID, row.Network, row.Address, row.DerivationPath, row.InvoiceID, row.Currency, row.CreatedAt,
	), nil
}

func (r *AddressRepository) ListByNetwork(ctx context.Context, network string) ([]*domain.AllocatedAddress, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var rows []addressRow
	if err := db.Where("network = ?", network).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}
	out := make([]*domain.AllocatedAddress, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.RestoreAllocatedAddress(
			row.ID, row.Network, row.Address, row.DerivationPath, row.InvoiceID, row.Currency, row.CreatedAt,
		))
	}
	return out, nil
}

func (r *AddressRepository) Save(ctx context.Context, addr *domain.AllocatedAddress) error {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return err
	}
	row := addressRow{
		ID:             addr.ID(),
		Network:        addr.Network(),
		Address:        addr.Address(),
		DerivationPath: addr.DerivationPath(),
		InvoiceID:      addr.InvoiceID(),
		Currency:       addr.Currency(),
		CreatedAt:      addr.CreatedAt().UTC(),
	}
	if err := db.Create(&row).Error; err != nil {
		return fmt.Errorf("insert address: %w", err)
	}
	return nil
}

func (r *AddressRepository) NextIndex(ctx context.Context, network string) (uint64, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return 0, err
	}

	var row counterRow
	err = db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("network = ?", network).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = counterRow{Network: network, NextIndex: 1}
		if err := db.Create(&row).Error; err != nil {
			return 0, fmt.Errorf("init counter: %w", err)
		}
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("load counter: %w", err)
	}

	idx := uint64(row.NextIndex)
	row.NextIndex++
	if err := db.Model(&counterRow{}).Where("network = ?", network).Update("next_index", row.NextIndex).Error; err != nil {
		return 0, fmt.Errorf("bump counter: %w", err)
	}
	return idx, nil
}
