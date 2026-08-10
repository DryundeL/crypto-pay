package blockchain

import "errors"

// Network identifies a supported chain family/network.
type Network string

const (
	NetworkEVMSepolia     Network = "evm:sepolia"
	NetworkBitcoinRegtest Network = "btc:regtest"
	NetworkBitcoinTestnet Network = "btc:testnet"
)

// ErrNotFound is returned when an address or watched tx cannot be found.
var ErrNotFound = errors.New("blockchain resource not found")

// ErrInvalid is returned for validation / business-rule failures on public facades.
var ErrInvalid = errors.New("invalid blockchain input")

// ErrInsufficientConfirmations is returned when Confirm is called before the network threshold.
var ErrInsufficientConfirmations = errors.New("insufficient confirmations")

// AddressAllocation is a public result of depositing address creation.
type AddressAllocation struct {
	Network        Network
	Address        string
	DerivationPath string
}

// WatchedAddress is a deposit address the scanner should monitor.
type WatchedAddress struct {
	Network   string
	Address   string
	InvoiceID string
	Currency  string
}

// DepositRef resolves a watched address to merchant metadata for observe/confirm.
type DepositRef struct {
	Network    string
	Address    string
	InvoiceID  string
	MerchantID string
	Currency   string
}

// RecordObservationInput is the public input for recording a chain observation.
type RecordObservationInput struct {
	MerchantID    string
	Network       string
	TxHash        string
	ToAddress     string
	Amount        string
	Currency      string
	Confirmations int
}

// ObservationView is a public projection of a watched transaction.
type ObservationView struct {
	ID            string
	Network       string
	TxHash        string
	ToAddress     string
	Amount        string
	Currency      string
	Confirmations int
	Status        string
}

// ConfirmTransactionInput is the public input for confirming a watched tx.
type ConfirmTransactionInput struct {
	MerchantID string
	Network    string
	TxHash     string
}
