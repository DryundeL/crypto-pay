package write

import (
	"context"
	"fmt"
)

// StubAddressProvider returns a deterministic fake deposit address.
// Replace with blockchain adapter when address allocation is implemented.
type StubAddressProvider struct{}

func NewStubAddressProvider() *StubAddressProvider {
	return &StubAddressProvider{}
}

func (p *StubAddressProvider) AllocateAddress(_ context.Context, network, _, invoiceID string) (string, error) {
	if network == "" {
		return "", fmt.Errorf("network is required")
	}
	if invoiceID == "" {
		return "", fmt.Errorf("invoice_id is required")
	}
	return fmt.Sprintf("stub:%s:%s", network, invoiceID), nil
}
