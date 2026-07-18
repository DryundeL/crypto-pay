package domain

import "fmt"

// Money is a simple value object for crypto amounts (string-based to avoid float).
// Precision rules will be enforced per asset later.
type Money struct {
	amount   string
	currency string
}

func NewMoney(amount, currency string) (Money, error) {
	if amount == "" {
		return Money{}, fmt.Errorf("%w: amount is required", ErrInvalidMoney)
	}
	if currency == "" {
		return Money{}, fmt.Errorf("%w: currency is required", ErrInvalidMoney)
	}
	return Money{amount: amount, currency: currency}, nil
}

func (m Money) Amount() string   { return m.amount }
func (m Money) Currency() string { return m.currency }
