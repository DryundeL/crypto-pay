package blockchain

import (
	"context"
	"sync"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/infrastructure/write"
	"github.com/DryundeL/crypto-pay/internal/platform/outbox"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// Module is the Blockchain bounded context composition root.
type Module struct {
	allocate *command.AllocateAddressHandler
	observe  *command.RecordObservationHandler
	confirm  *command.ConfirmTransactionHandler

	mu sync.Mutex
}

type Dependencies struct {
	DB            *gorm.DB
	Confirmations ports.ConfirmationPolicy
}

func NewModule(deps Dependencies) *Module {
	if deps.Confirmations == nil {
		panic("blockchain: Confirmations policy is required")
	}

	tx := transaction.NewGormManager(deps.DB)
	outboxStore := outbox.NewStore(deps.DB)
	publisher := write.NewOutboxPublisher(outboxStore)
	addrRepo := write.NewAddressRepository(deps.DB)
	txRepo := write.NewTransactionRepository(deps.DB)

	allocate := command.NewAllocateAddressHandler(addrRepo, tx)
	observe := command.NewRecordObservationHandler(txRepo, tx, publisher)
	confirm := command.NewConfirmTransactionHandler(txRepo, tx, publisher, deps.Confirmations)

	return &Module{
		allocate: allocate,
		observe:  observe,
		confirm:  confirm,
	}
}

func (m *Module) Name() string { return "blockchain" }

// SetPaymentNotifier wires sync payment callbacks after payment module is constructed.
func (m *Module) SetPaymentNotifier(n ports.PaymentNotifier) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.observe.SetPaymentNotifier(n)
	m.confirm.SetPaymentNotifier(n)
}

// AllocateAddress allocates (or returns existing) deposit address for an invoice.
func (m *Module) AllocateAddress(ctx context.Context, network, currency, invoiceID string) (AddressAllocation, error) {
	result, err := m.allocate.Handle(ctx, command.AllocateAddress{
		Network:   network,
		Currency:  currency,
		InvoiceID: invoiceID,
	})
	if err != nil {
		return AddressAllocation{}, err
	}
	return AddressAllocation{
		Network:        Network(result.Network),
		Address:        result.Address,
		DerivationPath: result.DerivationPath,
	}, nil
}

// AddressProvider adapts Module to invoice AddressProvider port (returns address string only).
func (m *Module) AddressProvider() *AddressProviderAdapter {
	return &AddressProviderAdapter{module: m}
}

// AddressProviderAdapter implements invoice's AllocateAddress(ctx, network, currency, invoiceID) (string, error).
type AddressProviderAdapter struct {
	module *Module
}

func (a *AddressProviderAdapter) AllocateAddress(ctx context.Context, network, currency, invoiceID string) (string, error) {
	alloc, err := a.module.AllocateAddress(ctx, network, currency, invoiceID)
	if err != nil {
		return "", err
	}
	return alloc.Address, nil
}
