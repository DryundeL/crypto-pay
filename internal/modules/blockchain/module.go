package blockchain

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/infrastructure/write"
	"github.com/DryundeL/crypto-pay/internal/platform/outbox"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// Module is the Blockchain bounded context composition root.
type Module struct {
	allocate *command.AllocateAddressHandler
	observe  *command.RecordObservationHandler
	confirm  *command.ConfirmTransactionHandler
	addrs    domain.AddressRepository
	deposits domain.DepositLookup
	cursors  domain.ScanCursorRepository

	mu sync.Mutex
}

type Dependencies struct {
	DB            *gorm.DB
	Confirmations ports.ConfirmationPolicy
	// EVMSepoliaXPub is account-level xpub at m/44'/60'/0' (required to allocate Sepolia addresses).
	EVMSepoliaXPub string
}

func NewModule(deps Dependencies) *Module {
	if deps.Confirmations == nil {
		panic("blockchain: Confirmations policy is required")
	}

	deriver, err := domain.NewHDAddressDeriver(deps.EVMSepoliaXPub)
	if err != nil {
		panic(fmt.Sprintf("blockchain: address deriver: %v", err))
	}

	tx := transaction.NewGormManager(deps.DB)
	outboxStore := outbox.NewStore(deps.DB)
	publisher := write.NewOutboxPublisher(outboxStore)
	addrRepo := write.NewAddressRepository(deps.DB)
	txRepo := write.NewTransactionRepository(deps.DB)
	depositLookup := write.NewDepositLookup(deps.DB)
	cursorRepo := write.NewScanCursorRepository(deps.DB)

	allocate := command.NewAllocateAddressHandler(addrRepo, tx, deriver)
	observe := command.NewRecordObservationHandler(txRepo, tx, publisher)
	confirm := command.NewConfirmTransactionHandler(txRepo, tx, publisher, deps.Confirmations)

	return &Module{
		allocate: allocate,
		observe:  observe,
		confirm:  confirm,
		addrs:    addrRepo,
		deposits: depositLookup,
		cursors:  cursorRepo,
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
		return AddressAllocation{}, mapFacadeError(err)
	}
	return AddressAllocation{
		Network:        Network(result.Network),
		Address:        result.Address,
		DerivationPath: result.DerivationPath,
	}, nil
}

// RecordObservation records (or refreshes) a watched chain transaction and notifies payment.
func (m *Module) RecordObservation(ctx context.Context, in RecordObservationInput) (ObservationView, error) {
	result, err := m.observe.Handle(ctx, command.RecordObservation{
		MerchantID:    in.MerchantID,
		Network:       in.Network,
		TxHash:        in.TxHash,
		ToAddress:     in.ToAddress,
		Amount:        in.Amount,
		Currency:      in.Currency,
		Confirmations: in.Confirmations,
	})
	if err != nil {
		return ObservationView{}, mapFacadeError(err)
	}
	return ObservationView{
		ID:            result.ID,
		Network:       result.Network,
		TxHash:        result.TxHash,
		ToAddress:     result.ToAddress,
		Amount:        result.Amount,
		Currency:      result.Currency,
		Confirmations: result.Confirmations,
		Status:        result.Status,
	}, nil
}

// ConfirmTransaction confirms a watched tx when confirmations meet policy and notifies payment.
func (m *Module) ConfirmTransaction(ctx context.Context, in ConfirmTransactionInput) (ObservationView, error) {
	result, err := m.confirm.Handle(ctx, command.ConfirmTransaction{
		MerchantID: in.MerchantID,
		Network:    in.Network,
		TxHash:     in.TxHash,
	})
	if err != nil {
		return ObservationView{}, mapFacadeError(err)
	}
	return ObservationView{
		ID:            result.ID,
		Network:       result.Network,
		TxHash:        result.TxHash,
		ToAddress:     result.ToAddress,
		Amount:        result.Amount,
		Currency:      result.Currency,
		Confirmations: result.Confirmations,
		Status:        result.Status,
	}, nil
}

// ListWatchedAddresses returns all deposit addresses for a network.
func (m *Module) ListWatchedAddresses(ctx context.Context, network string) ([]WatchedAddress, error) {
	if err := domain.ValidateNetwork(network); err != nil {
		return nil, mapFacadeError(err)
	}
	addrs, err := m.addrs.ListByNetwork(ctx, network)
	if err != nil {
		return nil, err
	}
	out := make([]WatchedAddress, 0, len(addrs))
	for _, a := range addrs {
		out = append(out, WatchedAddress{
			Network:   a.Network(),
			Address:   a.Address(),
			InvoiceID: a.InvoiceID(),
			Currency:  a.Currency(),
		})
	}
	return out, nil
}

// LookupDeposit resolves merchant metadata for a watched address.
func (m *Module) LookupDeposit(ctx context.Context, network, address string) (DepositRef, error) {
	if err := domain.ValidateNetwork(network); err != nil {
		return DepositRef{}, mapFacadeError(err)
	}
	ref, err := m.deposits.LookupByNetworkAddress(ctx, network, address)
	if err != nil {
		return DepositRef{}, mapFacadeError(err)
	}
	return DepositRef{
		Network:    ref.Network,
		Address:    ref.Address,
		InvoiceID:  ref.InvoiceID,
		MerchantID: ref.MerchantID,
		Currency:   ref.Currency,
	}, nil
}

// GetScanCursor returns the last processed block for a network.
func (m *Module) GetScanCursor(ctx context.Context, network string) (blockNumber uint64, found bool, err error) {
	return m.cursors.Get(ctx, network)
}

// SaveScanCursor upserts the last processed block for a network.
func (m *Module) SaveScanCursor(ctx context.Context, network string, blockNumber uint64) error {
	return m.cursors.Upsert(ctx, network, blockNumber)
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

func mapFacadeError(err error) error {
	switch {
	case errors.Is(err, domain.ErrAddressNotFound), errors.Is(err, domain.ErrTxNotFound):
		return ErrNotFound
	case errors.Is(err, domain.ErrInsufficientConfirmations):
		return ErrInsufficientConfirmations
	case errors.Is(err, domain.ErrInvalidBlockchain), errors.Is(err, domain.ErrInvalidTransition):
		return ErrInvalid
	default:
		return err
	}
}
