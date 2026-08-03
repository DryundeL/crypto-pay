package write

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

type DeliveryRepository struct {
	db *gorm.DB
}

func NewDeliveryRepository(db *gorm.DB) *DeliveryRepository {
	return &DeliveryRepository{db: db}
}

type deliveryRow struct {
	ID             string          `gorm:"column:id;primaryKey"`
	MerchantID     string          `gorm:"column:merchant_id"`
	EventName      string          `gorm:"column:event_name"`
	SourceID       string          `gorm:"column:source_id"`
	IdempotencyKey string          `gorm:"column:idempotency_key"`
	URL            string          `gorm:"column:url"`
	Payload        json.RawMessage `gorm:"column:payload;type:jsonb"`
	Status         string          `gorm:"column:status"`
	AttemptCount   int             `gorm:"column:attempt_count"`
	MaxAttempts    int             `gorm:"column:max_attempts"`
	NextAttemptAt  time.Time       `gorm:"column:next_attempt_at"`
	LastError      *string         `gorm:"column:last_error"`
	LastStatusCode *int            `gorm:"column:last_status_code"`
	DeliveredAt    *time.Time      `gorm:"column:delivered_at"`
	CreatedAt      time.Time       `gorm:"column:created_at"`
	UpdatedAt      time.Time       `gorm:"column:updated_at"`
}

func (deliveryRow) TableName() string { return "webhook_deliveries" }

func (r *DeliveryRepository) Save(ctx context.Context, d *domain.Delivery) error {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return err
	}

	row := deliveryRow{
		ID:             d.ID().String(),
		MerchantID:     d.MerchantID(),
		EventName:      d.EventName(),
		SourceID:       d.SourceID(),
		IdempotencyKey: d.IdempotencyKey(),
		URL:            d.URL(),
		Payload:        json.RawMessage(d.Payload()),
		Status:         d.Status().String(),
		AttemptCount:   d.AttemptCount(),
		MaxAttempts:    d.MaxAttempts(),
		NextAttemptAt:  d.NextAttemptAt().UTC(),
		DeliveredAt:    d.DeliveredAt(),
		CreatedAt:      d.CreatedAt().UTC(),
		UpdatedAt:      d.UpdatedAt().UTC(),
	}
	if d.LastError() != "" {
		e := d.LastError()
		row.LastError = &e
	}
	if d.AttemptCount() > 0 || d.LastStatusCode() != 0 {
		code := d.LastStatusCode()
		row.LastStatusCode = &code
	}

	if err := db.Save(&row).Error; err != nil {
		return fmt.Errorf("upsert webhook delivery: %w", err)
	}
	return nil
}

func (r *DeliveryRepository) FindByID(ctx context.Context, id domain.DeliveryID) (*domain.Delivery, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var row deliveryRow
	if err := db.Where("id = ?", id.String()).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDeliveryNotFound
		}
		return nil, fmt.Errorf("find webhook delivery: %w", err)
	}
	return mapDelivery(row), nil
}

func (r *DeliveryRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Delivery, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var row deliveryRow
	if err := db.Where("idempotency_key = ?", key).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDeliveryNotFound
		}
		return nil, fmt.Errorf("find webhook delivery by idempotency key: %w", err)
	}
	return mapDelivery(row), nil
}

func (r *DeliveryRepository) FindDue(ctx context.Context, now time.Time, limit int) ([]*domain.Delivery, error) {
	db, err := transaction.DB(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var rows []deliveryRow
	err = db.Where("status = ? AND next_attempt_at <= ?", domain.StatusPending.String(), now.UTC()).
		Order("next_attempt_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("find due webhook deliveries: %w", err)
	}

	out := make([]*domain.Delivery, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapDelivery(row))
	}
	return out, nil
}

func mapDelivery(row deliveryRow) *domain.Delivery {
	lastError := ""
	if row.LastError != nil {
		lastError = *row.LastError
	}
	lastStatusCode := 0
	if row.LastStatusCode != nil {
		lastStatusCode = *row.LastStatusCode
	}
	return domain.Restore(
		domain.DeliveryID(row.ID),
		row.MerchantID,
		row.EventName,
		row.SourceID,
		row.IdempotencyKey,
		row.URL,
		[]byte(row.Payload),
		domain.Status(row.Status),
		row.AttemptCount,
		row.MaxAttempts,
		row.NextAttemptAt,
		lastError,
		lastStatusCode,
		row.DeliveredAt,
		row.CreatedAt,
		row.UpdatedAt,
	)
}
