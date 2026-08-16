package outbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/DryundeL/crypto-pay/internal/platform/eventbus"
)

const (
	defaultRelayLimit = 100
	maxRelayLimit     = 500
)

type pendingProcessor interface {
	ProcessPending(ctx context.Context, limit int, fn func(context.Context, Message) error) (int, error)
}

type relayService struct {
	proc pendingProcessor
	pub  eventbus.Publisher
}

// NewRelayer publishes pending outbox rows through pub (in-process bus in the worker).
func NewRelayer(db *gorm.DB, pub eventbus.Publisher) Relayer {
	if db == nil {
		panic("outbox: db is required")
	}
	if pub == nil {
		panic("outbox: publisher is required")
	}
	return &relayService{proc: NewStore(db), pub: pub}
}

func (r *relayService) RelayPending(ctx context.Context, limit int) (int, error) {
	return r.proc.ProcessPending(ctx, clampRelayLimit(limit), func(ctx context.Context, msg Message) error {
		if err := r.pub.Publish(ctx, msg.EventName, msg.Payload); err != nil {
			return fmt.Errorf("publish %s: %w", msg.EventName, err)
		}
		return nil
	})
}

func clampRelayLimit(limit int) int {
	if limit <= 0 {
		return defaultRelayLimit
	}
	if limit > maxRelayLimit {
		return maxRelayLimit
	}
	return limit
}

// ProcessPending claims pending rows (FOR UPDATE SKIP LOCKED), runs fn, then marks sent.
// fn failure rolls back the claim so the row stays pending (at-least-once).
func (s *Store) ProcessPending(ctx context.Context, limit int, fn func(context.Context, Message) error) (int, error) {
	if fn == nil {
		return 0, fmt.Errorf("outbox: process fn is nil")
	}
	if s.db == nil {
		return 0, fmt.Errorf("outbox: db is nil")
	}

	processed := 0
	for processed < limit {
		claimed := false
		err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var row messageRow
			err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
				Where("status = ?", string(StatusPending)).
				Order("created_at ASC").
				Take(&row).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("claim pending outbox: %w", err)
			}

			claimed = true
			if err := fn(ctx, row.toMessage()); err != nil {
				return err
			}

			now := time.Now().UTC()
			res := tx.Model(&messageRow{}).
				Where("id = ? AND status = ?", row.ID, string(StatusPending)).
				Updates(map[string]any{
					"status":       string(StatusSent),
					"published_at": now,
				})
			if res.Error != nil {
				return fmt.Errorf("mark outbox sent: %w", res.Error)
			}
			if res.RowsAffected != 1 {
				return fmt.Errorf("mark outbox sent: expected 1 row, got %d", res.RowsAffected)
			}
			return nil
		})
		if err != nil {
			return processed, err
		}
		if !claimed {
			break
		}
		processed++
	}
	return processed, nil
}

func (r messageRow) toMessage() Message {
	return Message{
		ID:            r.ID,
		EventName:     r.EventName,
		AggregateType: r.AggregateType,
		AggregateID:   r.AggregateID,
		Payload:       r.Payload,
		Status:        Status(r.Status),
		CreatedAt:     r.CreatedAt,
		PublishedAt:   r.PublishedAt,
	}
}
