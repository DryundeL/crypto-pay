package query

import "context"

type WithdrawalReader interface {
	GetWithdrawal(ctx context.Context, withdrawalID, merchantID string) (WithdrawalDTO, error)
	ListWithdrawals(ctx context.Context, merchantID, status string, limit int) ([]WithdrawalDTO, error)
}

type GetWithdrawalHandler struct {
	reader WithdrawalReader
}

func NewGetWithdrawalHandler(reader WithdrawalReader) *GetWithdrawalHandler {
	return &GetWithdrawalHandler{reader: reader}
}

func (h *GetWithdrawalHandler) Handle(ctx context.Context, q GetWithdrawal) (WithdrawalDTO, error) {
	return h.reader.GetWithdrawal(ctx, q.WithdrawalID, q.MerchantID)
}
