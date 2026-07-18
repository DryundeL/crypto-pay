package query

import (
	"context"
	"fmt"
)

type APIKeyReadModel interface {
	ListAPIKeys(ctx context.Context, merchantID string) ([]APIKeyDTO, error)
}

type ListAPIKeysHandler struct {
	read APIKeyReadModel
}

func NewListAPIKeysHandler(read APIKeyReadModel) *ListAPIKeysHandler {
	return &ListAPIKeysHandler{read: read}
}

func (h *ListAPIKeysHandler) Handle(ctx context.Context, q ListAPIKeys) ([]APIKeyDTO, error) {
	if q.MerchantID == "" {
		return nil, fmt.Errorf("merchant_id is required")
	}
	return h.read.ListAPIKeys(ctx, q.MerchantID)
}
