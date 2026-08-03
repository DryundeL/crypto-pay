package query

import "context"

type DeliveryReader interface {
	GetDelivery(ctx context.Context, deliveryID, merchantID string) (DeliveryDTO, error)
	ListDeliveries(ctx context.Context, merchantID, status string, limit int) ([]DeliveryDTO, error)
}

type GetDeliveryHandler struct {
	reader DeliveryReader
}

func NewGetDeliveryHandler(reader DeliveryReader) *GetDeliveryHandler {
	return &GetDeliveryHandler{reader: reader}
}

func (h *GetDeliveryHandler) Handle(ctx context.Context, q GetDelivery) (DeliveryDTO, error) {
	return h.reader.GetDelivery(ctx, q.DeliveryID, q.MerchantID)
}
