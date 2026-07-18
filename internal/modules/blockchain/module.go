package blockchain

// Module is the Blockchain bounded context entrypoint (application wiring).
// Owns address allocation and scanners (EVM account-based + Bitcoin UTXO).
// Echo routes: internal/delivery/http only.
type Module struct{}

func NewModule() *Module { return &Module{} }

func (m *Module) Name() string { return "blockchain" }
