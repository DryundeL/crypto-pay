package command

import (
	"context"
	"errors"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/domain"
)

type ExpireDueHandler struct {
	repo   domain.Repository
	expire *ExpireInvoiceHandler
	now    func() time.Time
}

func NewExpireDueHandler(repo domain.Repository, expire *ExpireInvoiceHandler) *ExpireDueHandler {
	return &ExpireDueHandler{
		repo:   repo,
		expire: expire,
		now:    time.Now,
	}
}

func (h *ExpireDueHandler) Handle(ctx context.Context, cmd ExpireDue) (ExpireDueResult, error) {
	limit := cmd.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	ids, err := h.repo.ListExpiredPendingIDs(ctx, h.now().UTC(), limit)
	if err != nil {
		return ExpireDueResult{}, err
	}

	var result ExpireDueResult
	for _, id := range ids {
		err := h.expire.Handle(ctx, ExpireInvoice{InvoiceID: id.String()})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidTransition) ||
				errors.Is(err, domain.ErrInvoiceNotExpired) ||
				errors.Is(err, domain.ErrInvoiceNotFound) {
				result.Skipped++
				continue
			}
			return result, err
		}
		result.Expired++
	}
	return result, nil
}
