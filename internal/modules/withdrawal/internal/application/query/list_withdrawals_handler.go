package query

import "context"

type ListWithdrawalsHandler struct {
	reader WithdrawalReader
}

func NewListWithdrawalsHandler(reader WithdrawalReader) *ListWithdrawalsHandler {
	return &ListWithdrawalsHandler{reader: reader}
}

func (h *ListWithdrawalsHandler) Handle(ctx context.Context, q ListWithdrawals) ([]WithdrawalDTO, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return h.reader.ListWithdrawals(ctx, q.MerchantID, q.Status, limit)
}
