package domain

import "context"

// AddressRepository persists allocated deposit addresses.
type AddressRepository interface {
	FindByNetworkInvoice(ctx context.Context, network, invoiceID string) (*AllocatedAddress, error)
	FindByNetworkAddress(ctx context.Context, network, address string) (*AllocatedAddress, error)
	ListByNetwork(ctx context.Context, network string) ([]*AllocatedAddress, error)
	Save(ctx context.Context, addr *AllocatedAddress) error
	NextIndex(ctx context.Context, network string) (uint64, error)
}

// DepositLookup resolves a watched address to merchant + invoice metadata.
type DepositLookup interface {
	LookupByNetworkAddress(ctx context.Context, network, address string) (DepositRef, error)
}

// DepositRef is a cross-table projection for scanner matching.
type DepositRef struct {
	Network    string
	Address    string
	InvoiceID  string
	MerchantID string
	Currency   string
}

// ScanCursorRepository persists the last fully processed block per network.
type ScanCursorRepository interface {
	Get(ctx context.Context, network string) (blockNumber uint64, found bool, err error)
	Upsert(ctx context.Context, network string, blockNumber uint64) error
}

// TransactionRepository persists watched chain transactions.
type TransactionRepository interface {
	Save(ctx context.Context, tx *WatchedTransaction) error
	FindByNetworkTxHash(ctx context.Context, network, txHash string) (*WatchedTransaction, error)
}
