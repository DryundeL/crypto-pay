package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/crypto"
)

// Supported networks (mirrors public blockchain.Network constants).
const (
	NetworkEVMSepolia     = "evm:sepolia"
	NetworkBitcoinRegtest = "btc:regtest"
	NetworkBitcoinTestnet = "btc:testnet"
)

// AddressDeriver builds deposit addresses for a network.
type AddressDeriver interface {
	Derive(network string, index uint64) (address, derivationPath string, err error)
}

// HDAddressDeriver derives EVM addresses from an account-level xpub (m/44'/60'/0')
// via non-hardened path 0/{index}. BTC networks stay hash-placeholders until a UTXO scanner.
type HDAddressDeriver struct {
	evmAccount *hdkeychain.ExtendedKey
}

// NewHDAddressDeriver parses an account-level EVM xpub (m/44'/60'/0').
// Empty xpub is allowed; EVM derivation then fails with a clear error.
func NewHDAddressDeriver(evmSepoliaXPub string) (*HDAddressDeriver, error) {
	d := &HDAddressDeriver{}
	xpub := strings.TrimSpace(evmSepoliaXPub)
	if xpub == "" {
		return d, nil
	}
	key, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid EVM_SEPOLIA_XPUB: %v", ErrInvalidBlockchain, err)
	}
	if key.IsPrivate() {
		return nil, fmt.Errorf("%w: EVM_SEPOLIA_XPUB must be an extended public key", ErrInvalidBlockchain)
	}
	d.evmAccount = key
	return d, nil
}

// Derive implements AddressDeriver.
func (d *HDAddressDeriver) Derive(network string, index uint64) (address, derivationPath string, err error) {
	if err := ValidateNetwork(network); err != nil {
		return "", "", err
	}
	switch network {
	case NetworkEVMSepolia:
		return d.deriveEVM(index)
	case NetworkBitcoinRegtest, NetworkBitcoinTestnet:
		return deriveBTCPlaceholder(network, index)
	default:
		return "", "", fmt.Errorf("%w: unsupported network", ErrInvalidBlockchain)
	}
}

func (d *HDAddressDeriver) deriveEVM(index uint64) (string, string, error) {
	if d == nil || d.evmAccount == nil {
		return "", "", fmt.Errorf("%w: EVM_SEPOLIA_XPUB is required for evm:sepolia addresses", ErrInvalidBlockchain)
	}
	if index > uint64(^uint32(0)) {
		return "", "", fmt.Errorf("%w: derivation index out of range", ErrInvalidBlockchain)
	}
	change, err := d.evmAccount.Derive(0)
	if err != nil {
		return "", "", fmt.Errorf("derive change: %w", err)
	}
	child, err := change.Derive(uint32(index))
	if err != nil {
		return "", "", fmt.Errorf("derive index %d: %w", index, err)
	}
	pub, err := child.ECPubKey()
	if err != nil {
		return "", "", fmt.Errorf("ec pubkey: %w", err)
	}
	ecdsaPub, err := crypto.UnmarshalPubkey(pub.SerializeUncompressed())
	if err != nil {
		return "", "", fmt.Errorf("unmarshal pubkey: %w", err)
	}
	addr := crypto.PubkeyToAddress(*ecdsaPub)
	path := fmt.Sprintf("m/44'/60'/0'/0/%d", index)
	return strings.ToLower(addr.Hex()), path, nil
}

// deriveBTCPlaceholder keeps deterministic fake addresses until Bitcoin scanner lands.
// ponytail: hash placeholders — replace with BIP84 xpub derivation when btc scanner ships.
func deriveBTCPlaceholder(network string, index uint64) (string, string, error) {
	derivationPath := "m/0/" + strconv.FormatUint(index, 10)
	sum := sha256.Sum256([]byte(network + "|" + strconv.FormatUint(index, 10)))
	hexDigest := hex.EncodeToString(sum[:])
	switch network {
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
