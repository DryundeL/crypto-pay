package query

import (
	"context"
	"fmt"
)

type InvoiceReadModel interface {
	GetInvoice(ctx context.Context, invoiceID, merchantID string) (InvoiceDTO, error)
}

type GetInvoiceHandler struct {
	read InvoiceReadModel
}

func NewGetInvoiceHandler(read InvoiceReadModel) *GetInvoiceHandler {
	return &GetInvoiceHandler{read: read}
}

func (h *GetInvoiceHandler) Handle(ctx context.Context, q GetInvoice) (InvoiceDTO, error) {
	if q.InvoiceID == "" {
		return InvoiceDTO{}, fmt.Errorf("invoice_id is required")
	}
	if q.MerchantID == "" {
		return InvoiceDTO{}, fmt.Errorf("merchant_id is required")
	}
	return h.read.GetInvoice(ctx, q.InvoiceID, q.MerchantID)
}
