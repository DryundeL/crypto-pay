package read

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/domain"
)

type MerchantQueries struct {
	db *gorm.DB
}

func NewMerchantQueries(db *gorm.DB) *MerchantQueries {
	return &MerchantQueries{db: db}
}

type merchantDTORow struct {
	ID         string    `gorm:"column:id"`
	Name       string    `gorm:"column:name"`
	Status     string    `gorm:"column:status"`
	WebhookURL *string   `gorm:"column:webhook_url"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (merchantDTORow) TableName() string { return "merchants" }

type apiKeyDTORow struct {
	ID        string     `gorm:"column:id"`
	Name      string     `gorm:"column:name"`
	KeyPrefix string     `gorm:"column:key_prefix"`
	Status    string     `gorm:"column:status"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
}

func (apiKeyDTORow) TableName() string { return "merchant_api_keys" }

func (q *MerchantQueries) GetMerchant(ctx context.Context, merchantID string) (query.MerchantDTO, error) {
	var row merchantDTORow
	err := q.db.WithContext(ctx).
		Select("id", "name", "status", "webhook_url", "created_at", "updated_at").
		Where("id = ?", merchantID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.MerchantDTO{}, domain.ErrMerchantNotFound
		}
		return query.MerchantDTO{}, fmt.Errorf("query merchant: %w", err)
	}

	webhook := ""
	if row.WebhookURL != nil {
		webhook = *row.WebhookURL
	}

	return query.MerchantDTO{
		ID:         row.ID,
		Name:       row.Name,
		Status:     row.Status,
		WebhookURL: webhook,
		CreatedAt:  row.CreatedAt.UTC(),
		UpdatedAt:  row.UpdatedAt.UTC(),
	}, nil
}

func (q *MerchantQueries) ListAPIKeys(ctx context.Context, merchantID string) ([]query.APIKeyDTO, error) {
	var exists int64
	if err := q.db.WithContext(ctx).Table("merchants").Where("id = ?", merchantID).Count(&exists).Error; err != nil {
		return nil, fmt.Errorf("check merchant: %w", err)
	}
	if exists == 0 {
		return nil, domain.ErrMerchantNotFound
	}

	var rows []apiKeyDTORow
	err := q.db.WithContext(ctx).
		Select("id", "name", "key_prefix", "status", "created_at", "revoked_at").
		Where("merchant_id = ?", merchantID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("query api keys: %w", err)
	}

	out := make([]query.APIKeyDTO, 0, len(rows))
	for _, row := range rows {
		dto := query.APIKeyDTO{
			ID:        row.ID,
			Name:      row.Name,
			KeyPrefix: row.KeyPrefix,
			Status:    row.Status,
			CreatedAt: row.CreatedAt.UTC(),
		}
		if row.RevokedAt != nil {
			ts := row.RevokedAt.UTC()
			dto.RevokedAt = &ts
		}
		out = append(out, dto)
	}
	return out, nil
}
