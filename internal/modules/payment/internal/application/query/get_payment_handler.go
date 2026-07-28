package query

import (
	"context"
	"fmt"
)

type PaymentReadModel interface {
	GetPayment(ctx context.Context, paymentID, merchantID string) (PaymentDTO, error)
}

type GetPaymentHandler struct {
	read PaymentReadModel
}

func NewGetPaymentHandler(read PaymentReadModel) *GetPaymentHandler {
	return &GetPaymentHandler{read: read}
}

func (h *GetPaymentHandler) Handle(ctx context.Context, q GetPayment) (PaymentDTO, error) {
	if q.PaymentID == "" {
		return PaymentDTO{}, fmt.Errorf("payment_id is required")
	}
	if q.MerchantID == "" {
		return PaymentDTO{}, fmt.Errorf("merchant_id is required")
	}
	return h.read.GetPayment(ctx, q.PaymentID, q.MerchantID)
}
