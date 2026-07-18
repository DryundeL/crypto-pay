package query

import (
	"context"
	"fmt"
)

type MerchantReadModel interface {
	GetMerchant(ctx context.Context, merchantID string) (MerchantDTO, error)
}

type GetMerchantHandler struct {
	read MerchantReadModel
}

func NewGetMerchantHandler(read MerchantReadModel) *GetMerchantHandler {
	return &GetMerchantHandler{read: read}
}

func (h *GetMerchantHandler) Handle(ctx context.Context, q GetMerchant) (MerchantDTO, error) {
	if q.MerchantID == "" {
		return MerchantDTO{}, fmt.Errorf("merchant_id is required")
	}
	return h.read.GetMerchant(ctx, q.MerchantID)
}
