package query

import "context"

type ListDeliveriesHandler struct {
	reader DeliveryReader
}

func NewListDeliveriesHandler(reader DeliveryReader) *ListDeliveriesHandler {
	return &ListDeliveriesHandler{reader: reader}
}

func (h *ListDeliveriesHandler) Handle(ctx context.Context, q ListDeliveries) ([]DeliveryDTO, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return h.reader.ListDeliveries(ctx, q.MerchantID, q.Status, limit)
}
