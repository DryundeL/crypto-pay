package query

import "context"

type BalanceReader interface {
	ListBalances(ctx context.Context, merchantID string) ([]BalanceDTO, error)
}

type ListBalancesHandler struct {
	reader BalanceReader
}

func NewListBalancesHandler(reader BalanceReader) *ListBalancesHandler {
	return &ListBalancesHandler{reader: reader}
}

func (h *ListBalancesHandler) Handle(ctx context.Context, q ListBalances) ([]BalanceDTO, error) {
	return h.reader.ListBalances(ctx, q.MerchantID)
}
