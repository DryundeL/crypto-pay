package read

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/domain"
)

type DeliveryQueries struct {
	db *gorm.DB
}

func NewDeliveryQueries(db *gorm.DB) *DeliveryQueries {
	return &DeliveryQueries{db: db}
}

type deliveryDTORow struct {
	ID             string     `gorm:"column:id"`
	MerchantID     string     `gorm:"column:merchant_id"`
	EventName      string     `gorm:"column:event_name"`
	SourceID       string     `gorm:"column:source_id"`
	URL            string     `gorm:"column:url"`
	Status         string     `gorm:"column:status"`
	AttemptCount   int        `gorm:"column:attempt_count"`
	MaxAttempts    int        `gorm:"column:max_attempts"`
	NextAttemptAt  time.Time  `gorm:"column:next_attempt_at"`
	LastError      *string    `gorm:"column:last_error"`
	LastStatusCode *int       `gorm:"column:last_status_code"`
	DeliveredAt    *time.Time `gorm:"column:delivered_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (deliveryDTORow) TableName() string { return "webhook_deliveries" }

func (q *DeliveryQueries) GetDelivery(ctx context.Context, deliveryID, merchantID string) (query.DeliveryDTO, error) {
	var row deliveryDTORow
	err := q.db.WithContext(ctx).
		Where("id = ? AND merchant_id = ?", deliveryID, merchantID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.DeliveryDTO{}, domain.ErrDeliveryNotFound
		}
		return query.DeliveryDTO{}, fmt.Errorf("query webhook delivery: %w", err)
	}
	return mapDTO(row), nil
}

func (q *DeliveryQueries) ListDeliveries(ctx context.Context, merchantID, status string, limit int) ([]query.DeliveryDTO, error) {
	db := q.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit)
	if status != "" {
		db = db.Where("status = ?", status)
	}

	var rows []deliveryDTORow
	if err := db.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list webhook deliveries: %w", err)
	}

	out := make([]query.DeliveryDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapDTO(row))
	}
	return out, nil
}

func mapDTO(row deliveryDTORow) query.DeliveryDTO {
	dto := query.DeliveryDTO{
		ID:            row.ID,
		MerchantID:    row.MerchantID,
		EventName:     row.EventName,
		SourceID:      row.SourceID,
		URL:           row.URL,
		Status:        row.Status,
		AttemptCount:  row.AttemptCount,
		MaxAttempts:   row.MaxAttempts,
		NextAttemptAt: row.NextAttemptAt.UTC(),
		CreatedAt:     row.CreatedAt.UTC(),
		UpdatedAt:     row.UpdatedAt.UTC(),
	}
	if row.LastError != nil {
		dto.LastError = *row.LastError
	}
	if row.LastStatusCode != nil {
		code := *row.LastStatusCode
		dto.LastStatusCode = &code
	}
	if row.DeliveredAt != nil {
		t := row.DeliveredAt.UTC()
		dto.DeliveredAt = &t
	}
	return dto
}
