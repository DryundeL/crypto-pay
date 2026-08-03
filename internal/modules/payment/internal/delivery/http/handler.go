package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/payment/internal/domain"
)

type Handler struct {
	get *query.GetPaymentHandler
}

func NewHandler(get *query.GetPaymentHandler) *Handler {
	return &Handler{get: get}
}

func (h *Handler) GetPayment(c echo.Context) error {
	merchantID, ok := merchant.MerchantIDFromContext(c.Request().Context())
	if !ok || merchantID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	}

	dto, err := h.get.Handle(c.Request().Context(), query.GetPayment{
		PaymentID:  c.Param("id"),
		MerchantID: merchantID,
	})
	if err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusOK, paymentResponse{
		ID:            dto.ID,
		InvoiceID:     dto.InvoiceID,
		MerchantID:    dto.MerchantID,
		Status:        dto.Status,
		Network:       dto.Network,
		TxHash:        dto.TxHash,
		ToAddress:     dto.ToAddress,
		Amount:        dto.Amount,
		Currency:      dto.Currency,
		Confirmations: dto.Confirmations,
		CreatedAt:     dto.CreatedAt,
		UpdatedAt:     dto.UpdatedAt,
	})
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrPaymentNotFound):
		return c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidPayment),
		errors.Is(err, domain.ErrInvalidTransition):
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
