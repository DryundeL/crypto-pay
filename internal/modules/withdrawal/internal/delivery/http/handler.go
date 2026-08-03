package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/domain"
)

type Handler struct {
	request *command.RequestWithdrawalHandler
	get     *query.GetWithdrawalHandler
	list    *query.ListWithdrawalsHandler
}

func NewHandler(
	request *command.RequestWithdrawalHandler,
	get *query.GetWithdrawalHandler,
	list *query.ListWithdrawalsHandler,
) *Handler {
	return &Handler{request: request, get: get, list: list}
}

func (h *Handler) CreateWithdrawal(c echo.Context) error {
	merchantID, ok := merchant.MerchantIDFromContext(c.Request().Context())
	if !ok || merchantID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	}

	var req createWithdrawalRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid json body"})
	}

	result, err := h.request.Handle(c.Request().Context(), command.RequestWithdrawal{
		MerchantID:     merchantID,
		IdempotencyKey: req.IdempotencyKey,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Network:        req.Network,
		ToAddress:      req.ToAddress,
	})
	if err != nil {
		return mapError(c, err)
	}

	status := http.StatusCreated
	if !result.Created {
		status = http.StatusOK
	}
	return c.JSON(status, withdrawalResponse{
		ID:             result.ID,
		MerchantID:     result.MerchantID,
		IdempotencyKey: result.IdempotencyKey,
		Amount:         result.Amount,
		Currency:       result.Currency,
		Network:        result.Network,
		ToAddress:      result.ToAddress,
		Status:         result.Status,
		JournalID:      result.JournalID,
		CreatedAt:      result.CreatedAt,
		UpdatedAt:      result.UpdatedAt,
	})
}

func (h *Handler) GetWithdrawal(c echo.Context) error {
	merchantID, ok := merchant.MerchantIDFromContext(c.Request().Context())
	if !ok || merchantID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	}

	dto, err := h.get.Handle(c.Request().Context(), query.GetWithdrawal{
		WithdrawalID: c.Param("id"),
		MerchantID:   merchantID,
	})
	if err != nil {
		return mapError(c, err)
	}
	return c.JSON(http.StatusOK, toResponse(dto))
}

func (h *Handler) ListWithdrawals(c echo.Context) error {
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

	items, err := h.list.Handle(c.Request().Context(), query.ListWithdrawals{
		MerchantID: merchantID,
		Status:     c.QueryParam("status"),
		Limit:      limit,
	})
	if err != nil {
		return mapError(c, err)
	}

	out := make([]withdrawalResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toResponse(item))
	}
	return c.JSON(http.StatusOK, out)
}

func toResponse(dto query.WithdrawalDTO) withdrawalResponse {
	return withdrawalResponse{
		ID:             dto.ID,
		MerchantID:     dto.MerchantID,
		IdempotencyKey: dto.IdempotencyKey,
		Amount:         dto.Amount,
		Currency:       dto.Currency,
		Network:        dto.Network,
		ToAddress:      dto.ToAddress,
		Status:         dto.Status,
		TxHash:         dto.TxHash,
		JournalID:      dto.JournalID,
		CreatedAt:      dto.CreatedAt,
		UpdatedAt:      dto.UpdatedAt,
	}
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrWithdrawalNotFound):
		return c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, ledger.ErrInsufficientBalance):
		return c.JSON(http.StatusConflict, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidWithdrawal),
		errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, ledger.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
