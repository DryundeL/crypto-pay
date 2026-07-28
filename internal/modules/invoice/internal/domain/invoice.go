package domain

import (
	"fmt"
	"strings"
	"time"
)

// Invoice is the aggregate root of the Invoice bounded context.
type Invoice struct {
	id         InvoiceID
	merchantID string
	status     Status
	amount     Money
	network    string
	address    string
	txHash     string
	expiresAt  time.Time
	createdAt  time.Time
	updatedAt  time.Time
}

type CreateParams struct {
	ID         InvoiceID
	MerchantID string
	Amount     Money
	Network    string
	Address    string
	ExpiresAt  time.Time
	Now        time.Time
}

func Create(p CreateParams) (*Invoice, error) {
	if p.ID.IsZero() {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidInvoice)
	}
	merchantID := strings.TrimSpace(p.MerchantID)
	if merchantID == "" {
		return nil, fmt.Errorf("%w: merchant_id is required", ErrInvalidInvoice)
	}
	network := strings.TrimSpace(p.Network)
	if network == "" {
		return nil, fmt.Errorf("%w: network is required", ErrInvalidInvoice)
	}
	address := strings.TrimSpace(p.Address)
	if address == "" {
		return nil, fmt.Errorf("%w: address is required", ErrInvalidInvoice)
	}
	if p.ExpiresAt.IsZero() {
		return nil, fmt.Errorf("%w: expires_at is required", ErrInvalidInvoice)
	}
	now := p.Now.UTC()
	expiresAt := p.ExpiresAt.UTC()
	if !expiresAt.After(now) {
		return nil, fmt.Errorf("%w: expires_at must be in the future", ErrInvalidInvoice)
	}

	return &Invoice{
		id:         p.ID,
		merchantID: merchantID,
		status:     StatusPending,
		amount:     p.Amount,
		network:    network,
		address:    address,
		expiresAt:  expiresAt,
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

func Restore(
	id InvoiceID,
	merchantID string,
	status Status,
	amount Money,
	network, address, txHash string,
	expiresAt, createdAt, updatedAt time.Time,
) *Invoice {
	return &Invoice{
		id:         id,
		merchantID: merchantID,
		status:     status,
		amount:     amount,
		network:    network,
		address:    address,
		txHash:     txHash,
		expiresAt:  expiresAt,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

func (i *Invoice) ID() InvoiceID        { return i.id }
func (i *Invoice) MerchantID() string   { return i.merchantID }
func (i *Invoice) Status() Status       { return i.status }
func (i *Invoice) Amount() Money        { return i.amount }
func (i *Invoice) Network() string      { return i.network }
func (i *Invoice) Address() string      { return i.address }
func (i *Invoice) TxHash() string       { return i.txHash }
func (i *Invoice) ExpiresAt() time.Time { return i.expiresAt }
func (i *Invoice) CreatedAt() time.Time { return i.createdAt }
func (i *Invoice) UpdatedAt() time.Time { return i.updatedAt }

// Cancel transitions pending → cancelled.
func (i *Invoice) Cancel(now time.Time) error {
	if i.status != StatusPending {
		return fmt.Errorf("%w: cancel from %s", ErrInvalidTransition, i.status)
	}
	i.status = StatusCancelled
	i.updatedAt = now.UTC()
	return nil
}

// Expire transitions pending → expired when TTL has elapsed.
func (i *Invoice) Expire(now time.Time) error {
	if i.status != StatusPending {
		return fmt.Errorf("%w: expire from %s", ErrInvalidTransition, i.status)
	}
	now = now.UTC()
	if now.Before(i.expiresAt) {
		return ErrInvoiceNotExpired
	}
	i.status = StatusExpired
	i.updatedAt = now
	return nil
}

// MarkConfirming transitions pending → confirming.
func (i *Invoice) MarkConfirming(now time.Time) error {
	if i.status != StatusPending {
		return fmt.Errorf("%w: mark confirming from %s", ErrInvalidTransition, i.status)
	}
	i.status = StatusConfirming
	i.updatedAt = now.UTC()
	return nil
}

// MarkPaid transitions pending|confirming → paid.
func (i *Invoice) MarkPaid(txHash string, now time.Time) error {
	txHash = strings.TrimSpace(txHash)
	if txHash == "" {
		return fmt.Errorf("%w: tx_hash is required", ErrInvalidInvoice)
	}
	switch i.status {
	case StatusPending, StatusConfirming:
		// ok
	default:
		return fmt.Errorf("%w: mark paid from %s", ErrInvalidTransition, i.status)
	}
	i.status = StatusPaid
	i.txHash = txHash
	i.updatedAt = now.UTC()
	return nil
}
