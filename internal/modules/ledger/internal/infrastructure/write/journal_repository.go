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

type JournalRepository struct {
	db *gorm.DB
}

func NewJournalRepository(db *gorm.DB) *JournalRepository {
	return &JournalRepository{db: db}
}

type journalRow struct {
	ID             string    `gorm:"column:id;primaryKey"`
	IdempotencyKey string    `gorm:"column:idempotency_key"`
	MerchantID     string    `gorm:"column:merchant_id"`
	ReferenceType  string    `gorm:"column:reference_type"`
	ReferenceID    string    `gorm:"column:reference_id"`
	Amount         string    `gorm:"column:amount"`
	Currency       string    `gorm:"column:currency"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (journalRow) TableName() string { return "ledger_journals" }

type entryRow struct {
	ID        string    `gorm:"column:id;primaryKey"`
	JournalID string    `gorm:"column:journal_id"`
	AccountID string    `gorm:"column:account_id"`
	Side      string    `gorm:"column:side"`
	Amount    string    `gorm:"column:amount"`
	Currency  string    `gorm:"column:currency"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (entryRow) TableName() string { return "ledger_entries" }

func (r *JournalRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Journal, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var row journalRow
	if err := db.Where("idempotency_key = ?", key).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrJournalNotFound
		}
		return nil, fmt.Errorf("find journal: %w", err)
	}

	var entryRows []entryRow
	if err := db.Where("journal_id = ?", row.ID).Order("side ASC").Find(&entryRows).Error; err != nil {
		return nil, fmt.Errorf("find journal entries: %w", err)
	}

	entries := make([]domain.Entry, 0, len(entryRows))
	for _, er := range entryRows {
		amount, err := domain.NewMoney(er.Amount, er.Currency)
		if err != nil {
			return nil, fmt.Errorf("parse entry amount: %w", err)
		}
		side, err := domain.ParseEntrySide(er.Side)
		if err != nil {
			return nil, err
		}
		entries = append(entries, domain.RestoreEntry(
			domain.EntryID(er.ID),
			domain.AccountID(er.AccountID),
			side,
			amount,
		))
	}

	return domain.RestoreJournal(
		domain.JournalID(row.ID),
		row.IdempotencyKey,
		row.MerchantID,
		row.ReferenceType,
		row.ReferenceID,
		entries,
		row.CreatedAt,
	), nil
}

func (r *JournalRepository) Save(ctx context.Context, journal *domain.Journal) error {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return err
	}

	jrow := journalRow{
		ID:             journal.ID().String(),
		IdempotencyKey: journal.IdempotencyKey(),
		MerchantID:     journal.MerchantID(),
		ReferenceType:  journal.ReferenceType(),
		ReferenceID:    journal.ReferenceID(),
		Amount:         journal.Amount().Amount(),
		Currency:       journal.Currency(),
		CreatedAt:      journal.CreatedAt().UTC(),
	}
	if err := db.Create(&jrow).Error; err != nil {
		return fmt.Errorf("insert journal: %w", err)
	}

	for _, e := range journal.Entries() {
		erow := entryRow{
			ID:        e.ID().String(),
			JournalID: journal.ID().String(),
			AccountID: e.AccountID().String(),
			Side:      e.Side().String(),
			Amount:    e.Amount().Amount(),
			Currency:  e.Amount().Currency(),
			CreatedAt: journal.CreatedAt().UTC(),
		}
		if err := db.Create(&erow).Error; err != nil {
			return fmt.Errorf("insert entry: %w", err)
		}
	}
	return nil
}
