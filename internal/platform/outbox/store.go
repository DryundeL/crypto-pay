package outbox

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// Store implements Writer against outbox_messages.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Append(ctx context.Context, msg Message) error {
	db, err := transaction.DB(ctx, s.db)
	if err != nil {
		return err
	}

	row := messageRow{
		ID:            msg.ID,
		EventName:     msg.EventName,
		AggregateType: msg.AggregateType,
		AggregateID:   msg.AggregateID,
		Payload:       []byte(msg.Payload),
		Status:        string(msg.Status),
		CreatedAt:     msg.CreatedAt.UTC(),
	}

	if err := db.Create(&row).Error; err != nil {
		return fmt.Errorf("append outbox message: %w", err)
	}
	return nil
}

// WithTx returns a writer bound to an existing transaction handle.
func (s *Store) WithTx(tx *gorm.DB) Writer {
	return &Store{db: tx}
}

type messageRow struct {
	ID            string     `gorm:"column:id;primaryKey"`
	EventName     string     `gorm:"column:event_name"`
	AggregateType string     `gorm:"column:aggregate_type"`
	AggregateID   string     `gorm:"column:aggregate_id"`
	Payload       []byte     `gorm:"column:payload;type:jsonb"`
	Status        string     `gorm:"column:status"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
}

func (messageRow) TableName() string { return "outbox_messages" }
