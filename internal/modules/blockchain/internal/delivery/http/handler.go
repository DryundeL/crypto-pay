package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant"
	"github.com/DryundeL/crypto-pay/internal/modules/payment"
)

type Handler struct {
	observe *command.RecordObservationHandler
	confirm *command.ConfirmTransactionHandler
}

func NewHandler(
	observe *command.RecordObservationHandler,
	confirm *command.ConfirmTransactionHandler,
) *Handler {
	return &Handler{observe: observe, confirm: confirm}
}

func (h *Handler) RecordObservation(c echo.Context) error {
	merchantID, ok := merchant.MerchantIDFromContext(c.Request().Context())
	if !ok || merchantID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	}

	var req observedRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid json body"})
	}

	result, err := h.observe.Handle(c.Request().Context(), command.RecordObservation{
		MerchantID:    merchantID,
		Network:       req.Network,
		TxHash:        req.TxHash,
		ToAddress:     req.ToAddress,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Confirmations: req.Confirmations,
	})
	if err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusCreated, txResponse{
		ID:            result.ID,
		Network:       result.Network,
		TxHash:        result.TxHash,
		ToAddress:     result.ToAddress,
		Amount:        result.Amount,
		Currency:      result.Currency,
		Confirmations: result.Confirmations,
		Status:        result.Status,
		CreatedAt:     result.CreatedAt,
		UpdatedAt:     result.UpdatedAt,
	})
}

func (h *Handler) ConfirmTransaction(c echo.Context) error {
	merchantID, ok := merchant.MerchantIDFromContext(c.Request().Context())
	if !ok || merchantID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	}

	var req confirmedRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid json body"})
	}

	result, err := h.confirm.Handle(c.Request().Context(), command.ConfirmTransaction{
		MerchantID: merchantID,
		Network:    req.Network,
		TxHash:     req.TxHash,
	})
	if err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusOK, txResponse{
		ID:            result.ID,
		Network:       result.Network,
		TxHash:        result.TxHash,
		ToAddress:     result.ToAddress,
		Amount:        result.Amount,
		Currency:      result.Currency,
		Confirmations: result.Confirmations,
		Status:        result.Status,
		CreatedAt:     result.CreatedAt,
		UpdatedAt:     result.UpdatedAt,
	})
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrTxNotFound),
		errors.Is(err, domain.ErrAddressNotFound),
		errors.Is(err, payment.ErrNotFound):
		return c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidBlockchain),
		errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, payment.ErrInvalid):
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
