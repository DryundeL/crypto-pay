package write

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

type accountRow struct {
	ID        string    `gorm:"column:id;primaryKey"`
	OwnerType string    `gorm:"column:owner_type"`
	OwnerID   string    `gorm:"column:owner_id"`
	Kind      string    `gorm:"column:kind"`
	Currency  string    `gorm:"column:currency"`
	Balance   string    `gorm:"column:balance"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (accountRow) TableName() string { return "ledger_accounts" }

func (r *AccountRepository) FindByKey(
	ctx context.Context,
	ownerType domain.OwnerType,
	ownerID string,
	kind domain.AccountKind,
	currency string,
) (*domain.Account, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var row accountRow
	err = db.Where(
		"owner_type = ? AND owner_id = ? AND kind = ? AND currency = ?",
		ownerType.String(), ownerID, kind.String(), currency,
	).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("find account: %w", err)
	}
	return mapAccount(row)
}

func (r *AccountRepository) Save(ctx context.Context, account *domain.Account) error {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return err
	}

	row := accountRow{
		ID:        account.ID().String(),
		OwnerType: account.OwnerType().String(),
		OwnerID:   account.OwnerID(),
		Kind:      account.Kind().String(),
		Currency:  account.Currency(),
		Balance:   account.Balance().Amount(),
		CreatedAt: account.CreatedAt().UTC(),
		UpdatedAt: account.UpdatedAt().UTC(),
	}
	if err := db.Save(&row).Error; err != nil {
		return fmt.Errorf("upsert account: %w", err)
	}
	return nil
}

func mapAccount(row accountRow) (*domain.Account, error) {
	balance, err := domain.ParseStoredMoney(row.Balance, row.Currency)
	if err != nil {
		return nil, fmt.Errorf("parse account balance: %w", err)
	}
	return domain.RestoreAccount(
		domain.AccountID(row.ID),
		domain.OwnerType(row.OwnerType),
		row.OwnerID,
		domain.AccountKind(row.Kind),
		balance,
		row.CreatedAt,
		row.UpdatedAt,
	), nil
}
