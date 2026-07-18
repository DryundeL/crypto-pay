package ports

import "context"

// AddressProvider requests a unique deposit address for an invoice.
// Implemented by Blockchain bounded context adapter.
type AddressProvider interface {
	AllocateAddress(ctx context.Context, network, currency, invoiceID string) (address string, err error)
}
