package domain

import (
	"fmt"
	"strings"
	"time"
)

const DefaultMaxAttempts = 8

// Delivery is the aggregate root of the Webhook bounded context.
type Delivery struct {
	id             DeliveryID
	merchantID     string
	eventName      string
	sourceID       string
	idempotencyKey string
	url            string
	payload        []byte
	status         Status
	attemptCount   int
	maxAttempts    int
	nextAttemptAt  time.Time
	lastError      string
	lastStatusCode int
	deliveredAt    *time.Time
	createdAt      time.Time
	updatedAt      time.Time
}

type EnqueueParams struct {
	ID          DeliveryID
	MerchantID  string
	EventName   string
	SourceID    string
	URL         string
	Payload     []byte
	MaxAttempts int
	Now         time.Time
}

func Enqueue(p EnqueueParams) (*Delivery, error) {
	if p.ID.IsZero() {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidDelivery)
	}
	merchantID := strings.TrimSpace(p.MerchantID)
	if merchantID == "" {
		return nil, fmt.Errorf("%w: merchant_id is required", ErrInvalidDelivery)
	}
	eventName := strings.TrimSpace(p.EventName)
	if eventName == "" {
		return nil, fmt.Errorf("%w: event_name is required", ErrInvalidDelivery)
	}
	sourceID := strings.TrimSpace(p.SourceID)
	if sourceID == "" {
		return nil, fmt.Errorf("%w: source_id is required", ErrInvalidDelivery)
	}
	url := strings.TrimSpace(p.URL)
	if url == "" {
		return nil, fmt.Errorf("%w: url is required", ErrInvalidDelivery)
	}
	if len(p.Payload) == 0 {
		return nil, fmt.Errorf("%w: payload is required", ErrInvalidDelivery)
	}
	maxAttempts := p.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = DefaultMaxAttempts
	}

	now := p.Now.UTC()
	return &Delivery{
		id:             p.ID,
		merchantID:     merchantID,
		eventName:      eventName,
		sourceID:       sourceID,
		idempotencyKey: eventName + ":" + sourceID,
		url:            url,
		payload:        append([]byte(nil), p.Payload...),
		status:         StatusPending,
		attemptCount:   0,
		maxAttempts:    maxAttempts,
		nextAttemptAt:  now,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

func Restore(
	id DeliveryID,
	merchantID, eventName, sourceID, idempotencyKey, url string,
	payload []byte,
	status Status,
	attemptCount, maxAttempts int,
	nextAttemptAt time.Time,
	lastError string,
	lastStatusCode int,
	deliveredAt *time.Time,
	createdAt, updatedAt time.Time,
) *Delivery {
	return &Delivery{
		id:             id,
		merchantID:     merchantID,
		eventName:      eventName,
		sourceID:       sourceID,
		idempotencyKey: idempotencyKey,
		url:            url,
		payload:        append([]byte(nil), payload...),
		status:         status,
		attemptCount:   attemptCount,
		maxAttempts:    maxAttempts,
		nextAttemptAt:  nextAttemptAt,
		lastError:      lastError,
		lastStatusCode: lastStatusCode,
		deliveredAt:    deliveredAt,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (d *Delivery) ID() DeliveryID           { return d.id }
func (d *Delivery) MerchantID() string       { return d.merchantID }
func (d *Delivery) EventName() string        { return d.eventName }
func (d *Delivery) SourceID() string         { return d.sourceID }
func (d *Delivery) IdempotencyKey() string   { return d.idempotencyKey }
func (d *Delivery) URL() string              { return d.url }
func (d *Delivery) Payload() []byte          { return append([]byte(nil), d.payload...) }
func (d *Delivery) Status() Status           { return d.status }
func (d *Delivery) AttemptCount() int        { return d.attemptCount }
func (d *Delivery) MaxAttempts() int         { return d.maxAttempts }
func (d *Delivery) NextAttemptAt() time.Time { return d.nextAttemptAt }
func (d *Delivery) LastError() string        { return d.lastError }
func (d *Delivery) LastStatusCode() int      { return d.lastStatusCode }
func (d *Delivery) DeliveredAt() *time.Time  { return d.deliveredAt }
func (d *Delivery) CreatedAt() time.Time     { return d.createdAt }
func (d *Delivery) UpdatedAt() time.Time     { return d.updatedAt }

func (d *Delivery) MarkSent(statusCode int, now time.Time) error {
	if d.status != StatusPending {
		return fmt.Errorf("%w: cannot mark sent from %s", ErrInvalidTransition, d.status)
	}
	now = now.UTC()
	d.attemptCount++
	d.status = StatusSent
	d.lastStatusCode = statusCode
	d.lastError = ""
	d.deliveredAt = &now
	d.updatedAt = now
	return nil
}

// MarkAttemptFailed records a failed attempt. retryable keeps status pending with backoff;
// otherwise (or when max attempts reached) status becomes failed.
func (d *Delivery) MarkAttemptFailed(statusCode int, errMsg string, retryable bool, now time.Time) error {
	if d.status != StatusPending {
		return fmt.Errorf("%w: cannot mark failed attempt from %s", ErrInvalidTransition, d.status)
	}
	now = now.UTC()
	d.attemptCount++
	d.lastStatusCode = statusCode
	d.lastError = strings.TrimSpace(errMsg)
	d.updatedAt = now

	if !retryable || d.attemptCount >= d.maxAttempts {
		d.status = StatusFailed
		return nil
	}
	d.nextAttemptAt = now.Add(backoff(d.attemptCount))
	return nil
}

func backoff(attempt int) time.Duration {
	// 30s * 2^(attempt-1), cap 1h
	if attempt < 1 {
		attempt = 1
	}
	shift := attempt - 1
	if shift > 6 {
		shift = 6 // 30s * 64 = 1920s < 1h; then cap
	}
	d := 30 * time.Second * time.Duration(1<<uint(shift))
	if d > time.Hour {
		return time.Hour
	}
	return d
}
