package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant"
)

type Handler struct {
	listBalances *query.ListBalancesHandler
	listJournals *query.ListJournalsHandler
}

func NewHandler(
	listBalances *query.ListBalancesHandler,
	listJournals *query.ListJournalsHandler,
) *Handler {
	return &Handler{
		listBalances: listBalances,
		listJournals: listJournals,
	}
}

func (h *Handler) ListBalances(c echo.Context) error {
	merchantID, ok := merchant.MerchantIDFromContext(c.Request().Context())
	if !ok || merchantID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	}

	items, err := h.listBalances.Handle(c.Request().Context(), query.ListBalances{
		MerchantID: merchantID,
	})
	if err != nil {
		return mapError(c, err)
	}

	out := make([]balanceResponse, 0, len(items))
	for _, item := range items {
		out = append(out, balanceResponse{
			AccountID: item.AccountID,
			Currency:  item.Currency,
			Kind:      item.Kind,
			Balance:   item.Balance,
		})
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) ListJournals(c echo.Context) error {
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

	items, err := h.listJournals.Handle(c.Request().Context(), query.ListJournals{
		MerchantID: merchantID,
		Currency:   c.QueryParam("currency"),
		Limit:      limit,
	})
	if err != nil {
		return mapError(c, err)
	}

	out := make([]journalResponse, 0, len(items))
	for _, item := range items {
		out = append(out, journalResponse{
			ID:            item.ID,
			MerchantID:    item.MerchantID,
			ReferenceType: item.ReferenceType,
			ReferenceID:   item.ReferenceID,
			Amount:        item.Amount,
			Currency:      item.Currency,
			CreatedAt:     item.CreatedAt,
		})
	}
	return c.JSON(http.StatusOK, out)
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidMoney),
		errors.Is(err, domain.ErrInvalidJournal),
		errors.Is(err, domain.ErrInvalidAccount),
		errors.Is(err, domain.ErrInsufficientBalance):
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
