package domain

import (
	"fmt"
	"strings"
	"time"
)

// Payment is the aggregate root of the Payment bounded context.
type Payment struct {
	id            PaymentID
	invoiceID     string
	merchantID    string
	network       string
	txHash        string
	toAddress     string
	amount        string
	currency      string
	confirmations int
	status        Status
	createdAt     time.Time
	updatedAt     time.Time
}

type DetectParams struct {
	ID            PaymentID
	InvoiceID     string
	MerchantID    string
	Network       string
	TxHash        string
	ToAddress     string
	Amount        string
	Currency      string
	Confirmations int
	Now           time.Time
}

// Detect creates a payment in detected status.
func Detect(p DetectParams) (*Payment, error) {
	if p.ID.IsZero() {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidPayment)
	}
	invoiceID := strings.TrimSpace(p.InvoiceID)
	if invoiceID == "" {
		return nil, fmt.Errorf("%w: invoice_id is required", ErrInvalidPayment)
	}
	merchantID := strings.TrimSpace(p.MerchantID)
	if merchantID == "" {
		return nil, fmt.Errorf("%w: merchant_id is required", ErrInvalidPayment)
	}
	network := strings.TrimSpace(p.Network)
	if network == "" {
		return nil, fmt.Errorf("%w: network is required", ErrInvalidPayment)
	}
	txHash := strings.TrimSpace(p.TxHash)
	if txHash == "" {
		return nil, fmt.Errorf("%w: tx_hash is required", ErrInvalidPayment)
	}
	toAddress := strings.TrimSpace(p.ToAddress)
	if toAddress == "" {
		return nil, fmt.Errorf("%w: to_address is required", ErrInvalidPayment)
	}
	amount := strings.TrimSpace(p.Amount)
	if amount == "" {
		return nil, fmt.Errorf("%w: amount is required", ErrInvalidPayment)
	}
	currency := strings.TrimSpace(p.Currency)
	if currency == "" {
		return nil, fmt.Errorf("%w: currency is required", ErrInvalidPayment)
	}
	if p.Confirmations < 0 {
		return nil, fmt.Errorf("%w: confirmations must be >= 0", ErrInvalidPayment)
	}

	now := p.Now.UTC()
	return &Payment{
		id:            p.ID,
		invoiceID:     invoiceID,
		merchantID:    merchantID,
		network:       network,
		txHash:        txHash,
		toAddress:     toAddress,
		amount:        amount,
		currency:      currency,
		confirmations: p.Confirmations,
		status:        StatusDetected,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func Restore(
	id PaymentID,
	invoiceID, merchantID, network, txHash, toAddress, amount, currency string,
	confirmations int,
	status Status,
	createdAt, updatedAt time.Time,
) *Payment {
	return &Payment{
		id:            id,
		invoiceID:     invoiceID,
		merchantID:    merchantID,
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

func (p *Payment) ID() PaymentID        { return p.id }
func (p *Payment) InvoiceID() string    { return p.invoiceID }
func (p *Payment) MerchantID() string   { return p.merchantID }
func (p *Payment) Network() string      { return p.network }
func (p *Payment) TxHash() string       { return p.txHash }
func (p *Payment) ToAddress() string    { return p.toAddress }
func (p *Payment) Amount() string       { return p.amount }
func (p *Payment) Currency() string     { return p.currency }
func (p *Payment) Confirmations() int   { return p.confirmations }
func (p *Payment) Status() Status       { return p.status }
func (p *Payment) CreatedAt() time.Time { return p.createdAt }
func (p *Payment) UpdatedAt() time.Time { return p.updatedAt }

// StartConfirming transitions detected → confirming.
func (p *Payment) StartConfirming(now time.Time) error {
	if p.status != StatusDetected {
		return fmt.Errorf("%w: start confirming from %s", ErrInvalidTransition, p.status)
	}
	p.status = StatusConfirming
	p.updatedAt = now.UTC()
	return nil
}

// Confirm transitions detected|confirming → confirmed.
func (p *Payment) Confirm(now time.Time) error {
	switch p.status {
	case StatusDetected, StatusConfirming:
		// ok
	case StatusConfirmed:
		return nil // idempotent no-op at domain? Plan says Confirm from confirming|detected. Idempotent handled at app layer.
	default:
		return fmt.Errorf("%w: confirm from %s", ErrInvalidTransition, p.status)
	}
	p.status = StatusConfirmed
	p.updatedAt = now.UTC()
	return nil
}
