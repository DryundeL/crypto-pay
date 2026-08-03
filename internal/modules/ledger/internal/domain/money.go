package domain

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

var amountRE = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

// Money is a decimal amount + currency (string-based, no float).
type Money struct {
	amount   string
	currency string
	rat      *big.Rat
}

func NewMoney(amount, currency string) (Money, error) {
	amount = strings.TrimSpace(amount)
	currency = strings.TrimSpace(currency)
	if amount == "" {
		return Money{}, fmt.Errorf("%w: amount is required", ErrInvalidMoney)
	}
	if !amountRE.MatchString(amount) {
		return Money{}, fmt.Errorf("%w: amount must be a positive decimal", ErrInvalidMoney)
	}
	if currency == "" {
		return Money{}, fmt.Errorf("%w: currency is required", ErrInvalidMoney)
	}
	if len(currency) > 32 {
		return Money{}, fmt.Errorf("%w: currency too long", ErrInvalidMoney)
	}

	rat := new(big.Rat)
	if _, ok := rat.SetString(amount); !ok {
		return Money{}, fmt.Errorf("%w: cannot parse amount", ErrInvalidMoney)
	}
	if rat.Sign() <= 0 {
		return Money{}, fmt.Errorf("%w: amount must be positive", ErrInvalidMoney)
	}

	return Money{
		amount:   formatRat(rat),
		currency: currency,
		rat:      rat,
	}, nil
}

func ZeroMoney(currency string) (Money, error) {
	currency = strings.TrimSpace(currency)
	if currency == "" {
		return Money{}, fmt.Errorf("%w: currency is required", ErrInvalidMoney)
	}
	return Money{
		amount:   "0",
		currency: currency,
		rat:      big.NewRat(0, 1),
	}, nil
}

// ParseStoredMoney restores a persisted balance (zero allowed, negative rejected).
func ParseStoredMoney(amount, currency string) (Money, error) {
	amount = strings.TrimSpace(amount)
	currency = strings.TrimSpace(currency)
	if amount == "" {
		return Money{}, fmt.Errorf("%w: amount is required", ErrInvalidMoney)
	}
	if amount == "0" || amount == "0.0" {
		return ZeroMoney(currency)
	}
	if strings.HasPrefix(amount, "-") {
		return Money{}, fmt.Errorf("%w: negative balance not supported", ErrInvalidMoney)
	}
	return NewMoney(amount, currency)
}

func (m Money) Amount() string   { return m.amount }
func (m Money) Currency() string { return m.currency }
func (m Money) IsZero() bool     { return m.rat == nil || m.rat.Sign() == 0 }

func (m Money) Equal(other Money) bool {
	if m.currency != other.currency {
		return false
	}
	if m.rat == nil || other.rat == nil {
		return m.amount == other.amount
	}
	return m.rat.Cmp(other.rat) == 0
}

func (m Money) Add(other Money) (Money, error) {
	if err := m.sameCurrency(other); err != nil {
		return Money{}, err
	}
	sum := new(big.Rat).Add(m.rat, other.rat)
	return Money{amount: formatRat(sum), currency: m.currency, rat: sum}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if err := m.sameCurrency(other); err != nil {
		return Money{}, err
	}
	diff := new(big.Rat).Sub(m.rat, other.rat)
	return Money{amount: formatRat(diff), currency: m.currency, rat: diff}, nil
}

func (m Money) Cmp(other Money) (int, error) {
	if err := m.sameCurrency(other); err != nil {
		return 0, err
	}
	return m.rat.Cmp(other.rat), nil
}

func (m Money) sameCurrency(other Money) error {
	if m.currency != other.currency {
		return fmt.Errorf("%w: currency mismatch %s vs %s", ErrInvalidMoney, m.currency, other.currency)
	}
	if m.rat == nil || other.rat == nil {
		return fmt.Errorf("%w: uninitialized money", ErrInvalidMoney)
	}
	return nil
}

func formatRat(r *big.Rat) string {
	s := r.FloatString(18)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	if s == "" || s == "-" {
		return "0"
	}
	return s
}
