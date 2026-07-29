package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

// Supported networks (mirrors public blockchain.Network constants).
const (
	NetworkEVMSepolia     = "evm:sepolia"
	NetworkBitcoinRegtest = "btc:regtest"
	NetworkBitcoinTestnet = "btc:testnet"
)

// DeriveAddress builds a deterministic deposit address for (network, index, invoiceID).
// Not HD — hash-shaped placeholders suitable for local simulation until real keys land.
func DeriveAddress(network string, index uint64, invoiceID string) (address, derivationPath string, err error) {
	if invoiceID == "" {
		return "", "", fmt.Errorf("%w: invoice_id is required", ErrInvalidBlockchain)
	}
	if err := ValidateNetwork(network); err != nil {
		return "", "", err
	}
	derivationPath = "m/0/" + strconv.FormatUint(index, 10)

	sum := sha256.Sum256([]byte(network + "|" + strconv.FormatUint(index, 10) + "|" + invoiceID))
	hexDigest := hex.EncodeToString(sum[:])

	switch network {
	case NetworkEVMSepolia:
		return "0x" + hexDigest[:40], derivationPath, nil
	case NetworkBitcoinRegtest:
		return "bcrt1q" + hexDigest[:38], derivationPath, nil
	case NetworkBitcoinTestnet:
		return "tb1q" + hexDigest[:38], derivationPath, nil
	default:
		return "", "", fmt.Errorf("%w: unsupported network", ErrInvalidBlockchain)
	}
}

// ValidateNetwork returns ErrInvalidBlockchain for unknown networks.
func ValidateNetwork(network string) error {
	switch network {
	case NetworkEVMSepolia, NetworkBitcoinRegtest, NetworkBitcoinTestnet:
		return nil
	default:
		return fmt.Errorf("%w: unsupported network", ErrInvalidBlockchain)
	}
}
