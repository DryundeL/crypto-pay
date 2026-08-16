package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/query"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/domain"
)

type Handler struct {
	createMerchant      *command.CreateMerchantHandler
	createAPIKey        *command.CreateAPIKeyHandler
	revokeAPIKey        *command.RevokeAPIKeyHandler
	rotateWebhookSecret *command.RotateWebhookSecretHandler
	getMerchant         *query.GetMerchantHandler
	listAPIKeys         *query.ListAPIKeysHandler
}

func NewHandler(
	createMerchant *command.CreateMerchantHandler,
	createAPIKey *command.CreateAPIKeyHandler,
	revokeAPIKey *command.RevokeAPIKeyHandler,
	rotateWebhookSecret *command.RotateWebhookSecretHandler,
	getMerchant *query.GetMerchantHandler,
	listAPIKeys *query.ListAPIKeysHandler,
) *Handler {
	return &Handler{
		createMerchant:      createMerchant,
		createAPIKey:        createAPIKey,
		revokeAPIKey:        revokeAPIKey,
		rotateWebhookSecret: rotateWebhookSecret,
		getMerchant:         getMerchant,
		listAPIKeys:         listAPIKeys,
	}
}

func (h *Handler) CreateMerchant(c echo.Context) error {
	var req createMerchantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid json body"})
	}

	result, err := h.createMerchant.Handle(c.Request().Context(), command.CreateMerchant{
		Name:       req.Name,
		WebhookURL: req.WebhookURL,
	})
	if err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusCreated, merchantResponse{
		ID:            result.MerchantID,
		Name:          result.Name,
		Status:        result.Status,
		WebhookURL:    result.WebhookURL,
		CreatedAt:     result.CreatedAt,
		APIKeyID:      result.APIKeyID,
		KeyPrefix:     result.KeyPrefix,
		APIKey:        result.APIKey,
		WebhookSecret: result.WebhookSecret,
	})
}

func (h *Handler) GetMerchant(c echo.Context) error {
	dto, err := h.getMerchant.Handle(c.Request().Context(), query.GetMerchant{
		MerchantID: c.Param("id"),
	})
	if err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusOK, merchantResponse{
		ID:         dto.ID,
		Name:       dto.Name,
		Status:     dto.Status,
		WebhookURL: dto.WebhookURL,
		CreatedAt:  dto.CreatedAt,
		UpdatedAt:  dto.UpdatedAt,
	})
}

func (h *Handler) CreateAPIKey(c echo.Context) error {
	var req createAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid json body"})
	}

	result, err := h.createAPIKey.Handle(c.Request().Context(), command.CreateAPIKey{
		MerchantID: c.Param("id"),
		Name:       req.Name,
	})
	if err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusCreated, createAPIKeyResponse{
		ID:        result.APIKeyID,
		Name:      result.Name,
		KeyPrefix: result.KeyPrefix,
		APIKey:    result.APIKey,
		CreatedAt: result.CreatedAt,
	})
}

func (h *Handler) ListAPIKeys(c echo.Context) error {
	items, err := h.listAPIKeys.Handle(c.Request().Context(), query.ListAPIKeys{
		MerchantID: c.Param("id"),
	})
	if err != nil {
		return mapError(c, err)
	}

	out := make([]apiKeyResponse, 0, len(items))
	for _, item := range items {
		out = append(out, apiKeyResponse{
			ID:        item.ID,
			Name:      item.Name,
			KeyPrefix: item.KeyPrefix,
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
			RevokedAt: item.RevokedAt,
		})
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) RevokeAPIKey(c echo.Context) error {
	err := h.revokeAPIKey.Handle(c.Request().Context(), command.RevokeAPIKey{
		MerchantID: c.Param("id"),
		APIKeyID:   c.Param("keyId"),
	})
	if err != nil {
		return mapError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) RotateWebhookSecret(c echo.Context) error {
	result, err := h.rotateWebhookSecret.Handle(c.Request().Context(), command.RotateWebhookSecret{
		MerchantID: c.Param("id"),
	})
	if err != nil {
		return mapError(c, err)
	}
	return c.JSON(http.StatusOK, rotateWebhookSecretResponse{
		WebhookSecret: result.WebhookSecret,
	})
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrMerchantNotFound), errors.Is(err, domain.ErrAPIKeyNotFound):
		return c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidMerchant),
		errors.Is(err, domain.ErrInvalidAPIKey),
		errors.Is(err, domain.ErrAPIKeyRevoked),
		errors.Is(err, domain.ErrMerchantSuspended):
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
