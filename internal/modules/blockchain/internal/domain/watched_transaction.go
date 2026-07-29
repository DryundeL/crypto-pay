package domain

import (
	"fmt"
	"strings"
	"time"
)

// WatchedTransaction is a chain observation tracked by the blockchain BC.
type WatchedTransaction struct {
	id            string
	network       string
	txHash        string
	toAddress     string
	amount        string
	currency      string
	confirmations int
	status        TxStatus
	createdAt     time.Time
	updatedAt     time.Time
}

type ObserveParams struct {
	ID            string
	Network       string
	TxHash        string
	ToAddress     string
	Amount        string
	Currency      string
	Confirmations int
	Now           time.Time
}

func Observe(p ObserveParams) (*WatchedTransaction, error) {
	if strings.TrimSpace(p.ID) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidBlockchain)
	}
	if err := ValidateNetwork(p.Network); err != nil {
		return nil, err
	}
	if strings.TrimSpace(p.TxHash) == "" {
		return nil, fmt.Errorf("%w: tx_hash is required", ErrInvalidBlockchain)
	}
	if strings.TrimSpace(p.ToAddress) == "" {
		return nil, fmt.Errorf("%w: to_address is required", ErrInvalidBlockchain)
	}
	if strings.TrimSpace(p.Amount) == "" {
		return nil, fmt.Errorf("%w: amount is required", ErrInvalidBlockchain)
	}
	if strings.TrimSpace(p.Currency) == "" {
		return nil, fmt.Errorf("%w: currency is required", ErrInvalidBlockchain)
	}
	if p.Confirmations < 0 {
		return nil, fmt.Errorf("%w: confirmations must be >= 0", ErrInvalidBlockchain)
	}
	now := p.Now.UTC()
	return &WatchedTransaction{
		id:            p.ID,
		network:       p.Network,
		txHash:        strings.TrimSpace(p.TxHash),
		toAddress:     strings.TrimSpace(p.ToAddress),
		amount:        strings.TrimSpace(p.Amount),
		currency:      strings.TrimSpace(p.Currency),
		confirmations: p.Confirmations,
		status:        TxStatusObserved,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func RestoreWatchedTransaction(
	id, network, txHash, toAddress, amount, currency string,
	confirmations int,
	status TxStatus,
	createdAt, updatedAt time.Time,
) *WatchedTransaction {
	return &WatchedTransaction{
		id:            id,
		network:       network,
		txHash:        txHash,
		toAddress:     toAddress,
		amount:        amount,
		currency:      currency,
		confirmations: confirmations,
		status:        status,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (t *WatchedTransaction) ID() string           { return t.id }
func (t *WatchedTransaction) Network() string      { return t.network }
func (t *WatchedTransaction) TxHash() string       { return t.txHash }
func (t *WatchedTransaction) ToAddress() string    { return t.toAddress }
func (t *WatchedTransaction) Amount() string       { return t.amount }
func (t *WatchedTransaction) Currency() string     { return t.currency }
func (t *WatchedTransaction) Confirmations() int   { return t.confirmations }
func (t *WatchedTransaction) Status() TxStatus     { return t.status }
func (t *WatchedTransaction) CreatedAt() time.Time { return t.createdAt }
func (t *WatchedTransaction) UpdatedAt() time.Time { return t.updatedAt }

// Confirm transitions observed → confirmed (idempotent if already confirmed).
func (t *WatchedTransaction) Confirm(now time.Time) error {
	switch t.status {
	case TxStatusConfirmed:
		return nil
	case TxStatusObserved:
		t.status = TxStatusConfirmed
		t.updatedAt = now.UTC()
		return nil
	default:
		return fmt.Errorf("%w: confirm from %s", ErrInvalidTransition, t.status)
	}
}
