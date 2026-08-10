package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/domain"
)

type AllocateAddressHandler struct {
	repo    domain.AddressRepository
	tx      ports.TransactionManager
	deriver domain.AddressDeriver
	now     func() time.Time
}

func NewAllocateAddressHandler(
	repo domain.AddressRepository,
	tx ports.TransactionManager,
	deriver domain.AddressDeriver,
) *AllocateAddressHandler {
	return &AllocateAddressHandler{repo: repo, tx: tx, deriver: deriver, now: time.Now}
}

type AllocateAddressResult struct {
	Network        string
	Address        string
	DerivationPath string
}

func (h *AllocateAddressHandler) Handle(ctx context.Context, cmd AllocateAddress) (AllocateAddressResult, error) {
	if err := domain.ValidateNetwork(cmd.Network); err != nil {
		return AllocateAddressResult{}, err
	}
	if cmd.Currency == "" {
		return AllocateAddressResult{}, fmt.Errorf("%w: currency is required", domain.ErrInvalidBlockchain)
	}
	if cmd.InvoiceID == "" {
		return AllocateAddressResult{}, fmt.Errorf("%w: invoice_id is required", domain.ErrInvalidBlockchain)
	}

	existing, err := h.repo.FindByNetworkInvoice(ctx, cmd.Network, cmd.InvoiceID)
	if err == nil {
		return AllocateAddressResult{
			Network:        existing.Network(),
			Address:        existing.Address(),
			DerivationPath: existing.DerivationPath(),
		}, nil
	}
	if !errors.Is(err, domain.ErrAddressNotFound) {
		return AllocateAddressResult{}, err
	}

	var result AllocateAddressResult
	err = h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		existing, err := h.repo.FindByNetworkInvoice(ctx, cmd.Network, cmd.InvoiceID)
		if err == nil {
			result = AllocateAddressResult{
				Network:        existing.Network(),
				Address:        existing.Address(),
				DerivationPath: existing.DerivationPath(),
			}
			return nil
		}
		if !errors.Is(err, domain.ErrAddressNotFound) {
			return err
		}

		index, err := h.repo.NextIndex(ctx, cmd.Network)
		if err != nil {
			return err
		}
		if h.deriver == nil {
			return fmt.Errorf("%w: address deriver is required", domain.ErrInvalidBlockchain)
		}
		addr, path, err := h.deriver.Derive(cmd.Network, index)
		if err != nil {
			return err
		}
		now := h.now().UTC()
		allocated, err := domain.NewAllocatedAddress(
			uuid.NewString(), cmd.Network, addr, path, cmd.InvoiceID, cmd.Currency, now,
		)
		if err != nil {
			return err
		}
		if err := h.repo.Save(ctx, allocated); err != nil {
			return err
		}
		result = AllocateAddressResult{
			Network:        allocated.Network(),
			Address:        allocated.Address(),
			DerivationPath: allocated.DerivationPath(),
		}
		return nil
	})
	if err != nil {
		return AllocateAddressResult{}, err
	}
	return result, nil
}
