package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/domain"
)

const (
	defaultExpiresInSeconds = 30 * 60
	minExpiresInSeconds     = 60
	maxExpiresInSeconds     = 24 * 60 * 60
)

type CreateInvoiceHandler struct {
	repo    domain.Repository
	tx      ports.TransactionManager
	pub     ports.EventPublisher
	address ports.AddressProvider
	now     func() time.Time
}

func NewCreateInvoiceHandler(
	repo domain.Repository,
	tx ports.TransactionManager,
	pub ports.EventPublisher,
	address ports.AddressProvider,
) *CreateInvoiceHandler {
	return &CreateInvoiceHandler{
		repo:    repo,
		tx:      tx,
		pub:     pub,
		address: address,
		now:     time.Now,
	}
}

type CreateInvoiceResult struct {
	ID         string
	MerchantID string
	Status     string
	Amount     string
	Currency   string
	Network    string
	Address    string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (h *CreateInvoiceHandler) Handle(ctx context.Context, cmd CreateInvoice) (CreateInvoiceResult, error) {
	if err := validateNetwork(cmd.Network); err != nil {
		return CreateInvoiceResult{}, err
	}

	money, err := domain.NewMoney(cmd.Amount, cmd.Currency)
	if err != nil {
		return CreateInvoiceResult{}, err
	}

	expiresIn := cmd.ExpiresInSeconds
	if expiresIn == 0 {
		expiresIn = defaultExpiresInSeconds
	}
	if expiresIn < minExpiresInSeconds || expiresIn > maxExpiresInSeconds {
		return CreateInvoiceResult{}, fmt.Errorf(
			"%w: expires_in_seconds must be between %d and %d",
			domain.ErrInvalidInvoice, minExpiresInSeconds, maxExpiresInSeconds,
		)
	}

	id := domain.InvoiceID(uuid.NewString())
	addr, err := h.address.AllocateAddress(ctx, cmd.Network, cmd.Currency, id.String())
	if err != nil {
		return CreateInvoiceResult{}, fmt.Errorf("allocate address: %w", err)
	}

	var result CreateInvoiceResult
	err = h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		now := h.now().UTC()
		inv, err := domain.Create(domain.CreateParams{
			ID:         id,
			MerchantID: cmd.MerchantID,
			Amount:     money,
			Network:    cmd.Network,
			Address:    addr,
			ExpiresAt:  now.Add(time.Duration(expiresIn) * time.Second),
			Now:        now,
		})
		if err != nil {
			return err
		}

		if err := h.repo.Save(ctx, inv); err != nil {
			return fmt.Errorf("save invoice: %w", err)
		}

		if err := h.pub.Publish(ctx, "invoice", inv.ID().String(), invoiceCreatedEvent{
			InvoiceID:  inv.ID().String(),
			MerchantID: inv.MerchantID(),
			Amount:     inv.Amount().Amount(),
			Currency:   inv.Amount().Currency(),
			Network:    inv.Network(),
			OccurredAt: now,
		}); err != nil {
			return err
		}

		result = CreateInvoiceResult{
			ID:         inv.ID().String(),
			MerchantID: inv.MerchantID(),
			Status:     inv.Status().String(),
			Amount:     inv.Amount().Amount(),
			Currency:   inv.Amount().Currency(),
			Network:    inv.Network(),
			Address:    inv.Address(),
			ExpiresAt:  inv.ExpiresAt(),
			CreatedAt:  inv.CreatedAt(),
			UpdatedAt:  inv.UpdatedAt(),
		}
		return nil
	})
	if err != nil {
		return CreateInvoiceResult{}, err
	}
	return result, nil
}

func validateNetwork(network string) error {
	switch blockchain.Network(network) {
	case blockchain.NetworkEVMSepolia,
		blockchain.NetworkBitcoinRegtest,
		blockchain.NetworkBitcoinTestnet:
		return nil
	default:
		return fmt.Errorf("%w: unsupported network", domain.ErrInvalidInvoice)
	}
}
