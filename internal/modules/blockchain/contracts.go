package blockchain

// Network identifies a supported chain family/network.
type Network string

const (
	NetworkEVMSepolia   Network = "evm:sepolia"
	NetworkBitcoinRegtest Network = "btc:regtest"
	NetworkBitcoinTestnet Network = "btc:testnet"
)

// AddressAllocation is a public result of depositing address creation.
type AddressAllocation struct {
	Network Network
	Address string
	DerivationPath string
}
