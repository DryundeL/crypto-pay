package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/domain"
)

type Handler struct {
	get  *query.GetDeliveryHandler
	list *query.ListDeliveriesHandler
}

func NewHandler(get *query.GetDeliveryHandler, list *query.ListDeliveriesHandler) *Handler {
	return &Handler{get: get, list: list}
}

func (h *Handler) GetDelivery(c echo.Context) error {
	merchantID, ok := merchant.MerchantIDFromContext(c.Request().Context())
	if !ok || merchantID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	}

	dto, err := h.get.Handle(c.Request().Context(), query.GetDelivery{
		DeliveryID: c.Param("id"),
		MerchantID: merchantID,
	})
	if err != nil {
		return mapError(c, err)
	}
	return c.JSON(http.StatusOK, toResponse(dto))
}

func (h *Handler) ListDeliveries(c echo.Context) error {
	merchantID, ok := merchant.MerchantIDFromContext(c.Request().Context())
	if !ok || merchantID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	}

	limit := 0
	if raw := c.QueryParam("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid limit"})
		}
		limit = n
	}

	items, err := h.list.Handle(c.Request().Context(), query.ListDeliveries{
		MerchantID: merchantID,
		Status:     c.QueryParam("status"),
		Limit:      limit,
	})
	if err != nil {
		return mapError(c, err)
	}

	out := make([]deliveryResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toResponse(item))
	}
	return c.JSON(http.StatusOK, out)
}

func toResponse(dto query.DeliveryDTO) deliveryResponse {
	return deliveryResponse{
		ID:             dto.ID,
		MerchantID:     dto.MerchantID,
		EventName:      dto.EventName,
		SourceID:       dto.SourceID,
		URL:            dto.URL,
		Status:         dto.Status,
		AttemptCount:   dto.AttemptCount,
		MaxAttempts:    dto.MaxAttempts,
		NextAttemptAt:  dto.NextAttemptAt,
		LastError:      dto.LastError,
		LastStatusCode: dto.LastStatusCode,
		DeliveredAt:    dto.DeliveredAt,
		CreatedAt:      dto.CreatedAt,
		UpdatedAt:      dto.UpdatedAt,
	}
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrDeliveryNotFound):
		return c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidDelivery):
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
