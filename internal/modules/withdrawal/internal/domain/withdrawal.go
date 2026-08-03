package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var amountRE = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

// Withdrawal is the aggregate root of the Withdrawal bounded context.
type Withdrawal struct {
	id             WithdrawalID
	merchantID     string
	idempotencyKey string
	amount         string
	currency       string
	network        string
	toAddress      string
	status         Status
	txHash         string
	journalID      string
	createdAt      time.Time
	updatedAt      time.Time
}

type RequestParams struct {
	ID             WithdrawalID
	MerchantID     string
	IdempotencyKey string
	Amount         string
	Currency       string
	Network        string
	ToAddress      string
	JournalID      string
	Now            time.Time
}

func Request(p RequestParams) (*Withdrawal, error) {
	if p.ID.IsZero() {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidWithdrawal)
	}
	merchantID := strings.TrimSpace(p.MerchantID)
	if merchantID == "" {
		return nil, fmt.Errorf("%w: merchant_id is required", ErrInvalidWithdrawal)
	}
	key := strings.TrimSpace(p.IdempotencyKey)
	if key == "" {
		return nil, fmt.Errorf("%w: idempotency_key is required", ErrInvalidWithdrawal)
	}
	if len(key) > 128 {
		return nil, fmt.Errorf("%w: idempotency_key too long", ErrInvalidWithdrawal)
	}
	amount := strings.TrimSpace(p.Amount)
	if amount == "" || !amountRE.MatchString(amount) {
		return nil, fmt.Errorf("%w: amount must be a positive decimal", ErrInvalidWithdrawal)
	}
	if isZeroAmount(amount) {
		return nil, fmt.Errorf("%w: amount must be positive", ErrInvalidWithdrawal)
	}
	currency := strings.TrimSpace(p.Currency)
	if currency == "" {
		return nil, fmt.Errorf("%w: currency is required", ErrInvalidWithdrawal)
	}
	network := strings.TrimSpace(p.Network)
	if network == "" {
		return nil, fmt.Errorf("%w: network is required", ErrInvalidWithdrawal)
	}
	toAddress := strings.TrimSpace(p.ToAddress)
	if toAddress == "" {
		return nil, fmt.Errorf("%w: to_address is required", ErrInvalidWithdrawal)
	}
	journalID := strings.TrimSpace(p.JournalID)
	if journalID == "" {
		return nil, fmt.Errorf("%w: journal_id is required", ErrInvalidWithdrawal)
	}

	now := p.Now.UTC()
	return &Withdrawal{
		id:             p.ID,
		merchantID:     merchantID,
		idempotencyKey: key,
		amount:         amount,
		currency:       currency,
		network:        network,
		toAddress:      toAddress,
		status:         StatusRequested,
		journalID:      journalID,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

func Restore(
	id WithdrawalID,
	merchantID, idempotencyKey, amount, currency, network, toAddress string,
	status Status,
	txHash, journalID string,
	createdAt, updatedAt time.Time,
) *Withdrawal {
	return &Withdrawal{
		id:             id,
		merchantID:     merchantID,
		idempotencyKey: idempotencyKey,
		amount:         amount,
		currency:       currency,
		network:        network,
		toAddress:      toAddress,
		status:         status,
		txHash:         txHash,
		journalID:      journalID,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (w *Withdrawal) ID() WithdrawalID       { return w.id }
func (w *Withdrawal) MerchantID() string     { return w.merchantID }
func (w *Withdrawal) IdempotencyKey() string { return w.idempotencyKey }
func (w *Withdrawal) Amount() string         { return w.amount }
func (w *Withdrawal) Currency() string       { return w.currency }
func (w *Withdrawal) Network() string        { return w.network }
func (w *Withdrawal) ToAddress() string      { return w.toAddress }
func (w *Withdrawal) Status() Status         { return w.status }
func (w *Withdrawal) TxHash() string         { return w.txHash }
func (w *Withdrawal) JournalID() string      { return w.journalID }
func (w *Withdrawal) CreatedAt() time.Time   { return w.createdAt }
func (w *Withdrawal) UpdatedAt() time.Time   { return w.updatedAt }

func (w *Withdrawal) Complete(txHash string, now time.Time) error {
	if w.status != StatusRequested {
		return fmt.Errorf("%w: cannot complete from %s", ErrInvalidTransition, w.status)
	}
	txHash = strings.TrimSpace(txHash)
	if txHash == "" {
		return fmt.Errorf("%w: tx_hash is required", ErrInvalidWithdrawal)
	}
	w.status = StatusCompleted
	w.txHash = txHash
	w.updatedAt = now.UTC()
	return nil
}

func (w *Withdrawal) Reject(now time.Time) error {
	if w.status != StatusRequested {
		return fmt.Errorf("%w: cannot reject from %s", ErrInvalidTransition, w.status)
	}
	w.status = StatusRejected
	w.updatedAt = now.UTC()
	return nil
}

func isZeroAmount(amount string) bool {
	trimmed := strings.TrimLeft(strings.TrimRight(strings.TrimRight(amount, "0"), "."), "0")
	return trimmed == "" || trimmed == "."
}
