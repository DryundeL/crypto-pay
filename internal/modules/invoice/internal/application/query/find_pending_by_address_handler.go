package query

import (
	"context"
	"fmt"
)

type PendingByAddressReadModel interface {
	FindPendingByAddress(ctx context.Context, network, address string) (InvoiceDTO, error)
}

type FindPendingByAddressHandler struct {
	read PendingByAddressReadModel
}

func NewFindPendingByAddressHandler(read PendingByAddressReadModel) *FindPendingByAddressHandler {
	return &FindPendingByAddressHandler{read: read}
}

func (h *FindPendingByAddressHandler) Handle(ctx context.Context, q FindPendingByAddress) (InvoiceDTO, error) {
	if q.Network == "" {
		return InvoiceDTO{}, fmt.Errorf("network is required")
	}
	if q.Address == "" {
		return InvoiceDTO{}, fmt.Errorf("address is required")
	}
	return h.read.FindPendingByAddress(ctx, q.Network, q.Address)
}
