package domain

import (
	"fmt"
	"strings"
	"time"
)

const SystemOwnerID = "system"

type OwnerType string

const (
	OwnerSystem   OwnerType = "system"
	OwnerMerchant OwnerType = "merchant"
)

func (t OwnerType) String() string { return string(t) }

func (t OwnerType) Valid() bool {
	return t == OwnerSystem || t == OwnerMerchant
}

type AccountKind string

const (
	KindClearing  AccountKind = "clearing"
	KindAvailable AccountKind = "available"
)

func (k AccountKind) String() string { return string(k) }

func (k AccountKind) Valid() bool {
	return k == KindClearing || k == KindAvailable
}

// Account is a ledger account with a cached balance.
type Account struct {
	id        AccountID
	ownerType OwnerType
	ownerID   string
	kind      AccountKind
	currency  string
	balance   Money
	createdAt time.Time
	updatedAt time.Time
}

type CreateAccountParams struct {
	ID        AccountID
	OwnerType OwnerType
	OwnerID   string
	Kind      AccountKind
	Currency  string
	Now       time.Time
}

func CreateAccount(p CreateAccountParams) (*Account, error) {
	if p.ID.IsZero() {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidAccount)
	}
	if !p.OwnerType.Valid() {
		return nil, fmt.Errorf("%w: invalid owner_type", ErrInvalidAccount)
	}
	if !p.Kind.Valid() {
		return nil, fmt.Errorf("%w: invalid kind", ErrInvalidAccount)
	}

	ownerID := strings.TrimSpace(p.OwnerID)
	switch p.OwnerType {
	case OwnerSystem:
		if ownerID == "" {
			ownerID = SystemOwnerID
		}
		if p.Kind != KindClearing {
			return nil, fmt.Errorf("%w: system account must be clearing", ErrInvalidAccount)
		}
	case OwnerMerchant:
		if ownerID == "" {
			return nil, fmt.Errorf("%w: merchant owner_id is required", ErrInvalidAccount)
		}
		if p.Kind != KindAvailable {
			return nil, fmt.Errorf("%w: merchant account must be available", ErrInvalidAccount)
		}
	}

	zero, err := ZeroMoney(p.Currency)
	if err != nil {
		return nil, err
	}
	now := p.Now.UTC()
	return &Account{
		id:        p.ID,
		ownerType: p.OwnerType,
		ownerID:   ownerID,
		kind:      p.Kind,
		currency:  zero.Currency(),
		balance:   zero,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func RestoreAccount(
	id AccountID,
	ownerType OwnerType,
	ownerID string,
	kind AccountKind,
	balance Money,
	createdAt, updatedAt time.Time,
) *Account {
	return &Account{
		id:        id,
		ownerType: ownerType,
		ownerID:   ownerID,
		kind:      kind,
		currency:  balance.Currency(),
		balance:   balance,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (a *Account) ID() AccountID        { return a.id }
func (a *Account) OwnerType() OwnerType { return a.ownerType }
func (a *Account) OwnerID() string      { return a.ownerID }
func (a *Account) Kind() AccountKind    { return a.kind }
func (a *Account) Currency() string     { return a.currency }
func (a *Account) Balance() Money       { return a.balance }
func (a *Account) CreatedAt() time.Time { return a.createdAt }
func (a *Account) UpdatedAt() time.Time { return a.updatedAt }

// ApplyEntry updates cached balance using account-kind polarity:
// available (liability): credit +, debit −
// clearing (asset): debit +, credit −
func (a *Account) ApplyEntry(side EntrySide, amount Money, now time.Time) error {
	if !side.Valid() {
		return fmt.Errorf("%w: invalid side", ErrInvalidAccount)
	}
	if amount.Currency() != a.currency {
		return fmt.Errorf("%w: currency mismatch", ErrInvalidAccount)
	}

	var (
		next Money
		err  error
	)
	switch a.kind {
	case KindAvailable:
		if side == SideCredit {
			next, err = a.balance.Add(amount)
		} else {
			next, err = a.balance.Sub(amount)
		}
	case KindClearing:
		if side == SideDebit {
			next, err = a.balance.Add(amount)
		} else {
			next, err = a.balance.Sub(amount)
		}
	default:
		return fmt.Errorf("%w: unknown kind", ErrInvalidAccount)
	}
	if err != nil {
		return err
	}

	// v1: merchant available must not go negative (future withdrawals).
	if a.kind == KindAvailable && next.rat.Sign() < 0 {
		return ErrInsufficientBalance
	}

	a.balance = next
	a.updatedAt = now.UTC()
	return nil
}
