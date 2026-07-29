package domain

import "context"

// AddressRepository persists allocated deposit addresses.
type AddressRepository interface {
	FindByNetworkInvoice(ctx context.Context, network, invoiceID string) (*AllocatedAddress, error)
	Save(ctx context.Context, addr *AllocatedAddress) error
	NextIndex(ctx context.Context, network string) (uint64, error)
}

// TransactionRepository persists watched chain transactions.
type TransactionRepository interface {
	Save(ctx context.Context, tx *WatchedTransaction) error
	FindByNetworkTxHash(ctx context.Context, network, txHash string) (*WatchedTransaction, error)
}
