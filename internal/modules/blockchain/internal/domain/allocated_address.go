package domain

import (
	"fmt"
	"strings"
	"time"
)

// AllocatedAddress is a persisted deposit address binding.
type AllocatedAddress struct {
	id             string
	network        string
	address        string
	derivationPath string
	invoiceID      string
	currency       string
	createdAt      time.Time
}

func NewAllocatedAddress(
	id, network, address, derivationPath, invoiceID, currency string,
	now time.Time,
) (*AllocatedAddress, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidBlockchain)
	}
	if err := ValidateNetwork(network); err != nil {
		return nil, err
	}
	if strings.TrimSpace(address) == "" {
		return nil, fmt.Errorf("%w: address is required", ErrInvalidBlockchain)
	}
	if strings.TrimSpace(derivationPath) == "" {
		return nil, fmt.Errorf("%w: derivation_path is required", ErrInvalidBlockchain)
	}
	if strings.TrimSpace(invoiceID) == "" {
		return nil, fmt.Errorf("%w: invoice_id is required", ErrInvalidBlockchain)
	}
	if strings.TrimSpace(currency) == "" {
		return nil, fmt.Errorf("%w: currency is required", ErrInvalidBlockchain)
	}
	return &AllocatedAddress{
		id:             id,
		network:        network,
		address:        address,
		derivationPath: derivationPath,
		invoiceID:      invoiceID,
		currency:       currency,
		createdAt:      now.UTC(),
	}, nil
}

func RestoreAllocatedAddress(
	id, network, address, derivationPath, invoiceID, currency string,
	createdAt time.Time,
) *AllocatedAddress {
	return &AllocatedAddress{
		id:             id,
		network:        network,
		address:        address,
		derivationPath: derivationPath,
		invoiceID:      invoiceID,
		currency:       currency,
		createdAt:      createdAt,
	}
}

func (a *AllocatedAddress) ID() string             { return a.id }
func (a *AllocatedAddress) Network() string        { return a.network }
func (a *AllocatedAddress) Address() string        { return a.address }
func (a *AllocatedAddress) DerivationPath() string { return a.derivationPath }
func (a *AllocatedAddress) InvoiceID() string      { return a.invoiceID }
func (a *AllocatedAddress) Currency() string       { return a.currency }
func (a *AllocatedAddress) CreatedAt() time.Time   { return a.createdAt }
