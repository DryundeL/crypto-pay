package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant"
)

type Handler struct {
	create *command.CreateInvoiceHandler
	cancel *command.CancelInvoiceHandler
	get    *query.GetInvoiceHandler
}

func NewHandler(
	create *command.CreateInvoiceHandler,
	cancel *command.CancelInvoiceHandler,
	get *query.GetInvoiceHandler,
) *Handler {
	return &Handler{
		create: create,
		cancel: cancel,
		get:    get,
	}
}

func (h *Handler) CreateInvoice(c echo.Context) error {
	merchantID, ok := merchant.MerchantIDFromContext(c.Request().Context())
	if !ok || merchantID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	}

	var req createInvoiceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid json body"})
	}

	result, err := h.create.Handle(c.Request().Context(), command.CreateInvoice{
		MerchantID:       merchantID,
		Amount:           req.Amount,
		Currency:         req.Currency,
		Network:          req.Network,
		ExpiresInSeconds: req.ExpiresInSeconds,
	})
	if err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusCreated, invoiceResponse{
		ID:         result.ID,
		MerchantID: result.MerchantID,
		Status:     result.Status,
		Amount:     result.Amount,
		Currency:   result.Currency,
		Network:    result.Network,
		Address:    result.Address,
		ExpiresAt:  result.ExpiresAt,
		CreatedAt:  result.CreatedAt,
		UpdatedAt:  result.UpdatedAt,
	})
}

func (h *Handler) GetInvoice(c echo.Context) error {
	merchantID, ok := merchant.MerchantIDFromContext(c.Request().Context())
	if !ok || merchantID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	}

	dto, err := h.get.Handle(c.Request().Context(), query.GetInvoice{
		InvoiceID:  c.Param("id"),
		MerchantID: merchantID,
	})
	if err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusOK, invoiceResponse{
		ID:         dto.ID,
		MerchantID: dto.MerchantID,
		Status:     dto.Status,
		Amount:     dto.Amount,
		Currency:   dto.Currency,
		Network:    dto.Network,
		Address:    dto.Address,
		TxHash:     dto.TxHash,
		ExpiresAt:  dto.ExpiresAt,
		CreatedAt:  dto.CreatedAt,
		UpdatedAt:  dto.UpdatedAt,
	})
}

func (h *Handler) CancelInvoice(c echo.Context) error {
	merchantID, ok := merchant.MerchantIDFromContext(c.Request().Context())
	if !ok || merchantID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	}

	result, err := h.cancel.Handle(c.Request().Context(), command.CancelInvoice{
		InvoiceID:  c.Param("id"),
		MerchantID: merchantID,
	})
	if err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusOK, invoiceResponse{
		ID:         result.ID,
		MerchantID: result.MerchantID,
		Status:     result.Status,
		Amount:     result.Amount,
		Currency:   result.Currency,
		Network:    result.Network,
		Address:    result.Address,
		TxHash:     result.TxHash,
		ExpiresAt:  result.ExpiresAt,
		CreatedAt:  result.CreatedAt,
		UpdatedAt:  result.UpdatedAt,
	})
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrInvoiceNotFound):
		return c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidMoney),
		errors.Is(err, domain.ErrInvalidInvoice),
		errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrInvoiceNotExpired):
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
