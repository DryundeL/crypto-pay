package write

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

type MerchantRepository struct {
	db *gorm.DB
}

func NewMerchantRepository(db *gorm.DB) *MerchantRepository {
	return &MerchantRepository{db: db}
}

type merchantRow struct {
	ID            string    `gorm:"column:id;primaryKey"`
	Name          string    `gorm:"column:name"`
	Status        string    `gorm:"column:status"`
	WebhookURL    *string   `gorm:"column:webhook_url"`
	WebhookSecret string    `gorm:"column:webhook_secret"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (merchantRow) TableName() string { return "merchants" }

type apiKeyRow struct {
	ID         string     `gorm:"column:id;primaryKey"`
	MerchantID string     `gorm:"column:merchant_id"`
	Name       string     `gorm:"column:name"`
	KeyPrefix  string     `gorm:"column:key_prefix"`
	KeyHash    string     `gorm:"column:key_hash"`
	Status     string     `gorm:"column:status"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	RevokedAt  *time.Time `gorm:"column:revoked_at"`
}

func (apiKeyRow) TableName() string { return "merchant_api_keys" }

func (r *MerchantRepository) Save(ctx context.Context, merchant *domain.Merchant) error {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return err
	}

	var webhook *string
	if merchant.WebhookURL() != "" {
		u := merchant.WebhookURL()
		webhook = &u
	}

	row := merchantRow{
		ID:            merchant.ID().String(),
		Name:          merchant.Name(),
		Status:        merchant.Status().String(),
		WebhookURL:    webhook,
		WebhookSecret: merchant.WebhookSecret(),
		CreatedAt:     merchant.CreatedAt().UTC(),
		UpdatedAt:     merchant.UpdatedAt().UTC(),
	}

	if err := db.Save(&row).Error; err != nil {
		return fmt.Errorf("upsert merchant: %w", err)
	}

	for _, key := range merchant.APIKeys() {
		keyRow := apiKeyRow{
			ID:         key.ID().String(),
			MerchantID: merchant.ID().String(),
			Name:       key.Name(),
			KeyPrefix:  key.KeyPrefix(),
			KeyHash:    key.KeyHash(),
			Status:     key.Status().String(),
			CreatedAt:  key.CreatedAt().UTC(),
			RevokedAt:  key.RevokedAt(),
		}
		if err := db.Save(&keyRow).Error; err != nil {
			return fmt.Errorf("upsert api key: %w", err)
		}
	}

	return nil
}

func (r *MerchantRepository) FindByID(ctx context.Context, id domain.MerchantID) (*domain.Merchant, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var row merchantRow
	if err := db.Where("id = ?", id.String()).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMerchantNotFound
		}
		return nil, fmt.Errorf("load merchant: %w", err)
	}

	var keyRows []apiKeyRow
	if err := db.Where("merchant_id = ?", id.String()).Order("created_at ASC").Find(&keyRows).Error; err != nil {
		return nil, fmt.Errorf("load api keys: %w", err)
	}

	keys := make([]*domain.APIKey, 0, len(keyRows))
	for _, kr := range keyRows {
		keys = append(keys, domain.RestoreAPIKey(
			domain.APIKeyID(kr.ID),
			kr.Name,
			kr.KeyPrefix,
			kr.KeyHash,
			domain.APIKeyStatus(kr.Status),
			kr.CreatedAt,
			kr.RevokedAt,
		))
	}

	webhook := ""
	if row.WebhookURL != nil {
		webhook = *row.WebhookURL
	}

	return domain.RestoreMerchant(
		domain.MerchantID(row.ID),
		row.Name,
		domain.Status(row.Status),
		webhook,
		row.WebhookSecret,
		keys,
		row.CreatedAt,
		row.UpdatedAt,
	), nil
}
