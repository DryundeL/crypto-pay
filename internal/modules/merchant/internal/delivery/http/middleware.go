package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/authctx"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/domain"
)

const (
	headerAPIKey     = "X-API-Key"
	maxAPIKeyLength  = 256
	bearerPrefix     = "Bearer "
	apiKeyAuthPrefix = "ApiKey "
)

// APIKeyAuth authenticates requests by hashing the presented API key and
// looking up merchant_api_keys.key_hash. On success stores principal in context.
//
// If route param "id" is present, it must match the authenticated merchant
// (merchant-owned routes like /merchants/:id).
func APIKeyAuth(hasher ports.APIKeyGenerator, auth ports.APIKeyAuthenticator) echo.MiddlewareFunc {
	return apiKeyAuth(hasher, auth, true)
}

// APIKeyAuthOnly authenticates the API key without binding route param "id"
// to the merchant. Use for other modules (e.g. /invoices/:id).
func APIKeyAuthOnly(hasher ports.APIKeyGenerator, auth ports.APIKeyAuthenticator) echo.MiddlewareFunc {
	return apiKeyAuth(hasher, auth, false)
}

func apiKeyAuth(hasher ports.APIKeyGenerator, auth ports.APIKeyAuthenticator, bindRouteMerchantID bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			raw, ok := extractAPIKey(c)
			if !ok {
				return c.JSON(http.StatusUnauthorized, errorResponse{Error: "missing api key"})
			}
			if len(raw) > maxAPIKeyLength {
				return c.JSON(http.StatusUnauthorized, errorResponse{Error: "invalid api key"})
			}

			principal, err := auth.AuthenticateByHash(c.Request().Context(), hasher.Hash(raw))
			if err != nil {
				return mapAuthError(c, err)
			}

			if bindRouteMerchantID {
				if routeMerchantID := c.Param("id"); routeMerchantID != "" && routeMerchantID != principal.MerchantID {
					return c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
				}
			}

			ctx := authctx.WithAuth(c.Request().Context(), principal.MerchantID, principal.APIKeyID)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func extractAPIKey(c echo.Context) (string, bool) {
	if key := strings.TrimSpace(c.Request().Header.Get(headerAPIKey)); key != "" {
		return key, true
	}

	authz := strings.TrimSpace(c.Request().Header.Get(echo.HeaderAuthorization))
	switch {
	case strings.HasPrefix(authz, bearerPrefix):
		key := strings.TrimSpace(strings.TrimPrefix(authz, bearerPrefix))
		return key, key != ""
	case strings.HasPrefix(authz, apiKeyAuthPrefix):
		key := strings.TrimSpace(strings.TrimPrefix(authz, apiKeyAuthPrefix))
		return key, key != ""
	default:
		return "", false
	}
}

func mapAuthError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrAPIKeyNotFound),
		errors.Is(err, domain.ErrInvalidAPIKey),
		errors.Is(err, domain.ErrAPIKeyRevoked):
		// Uniform message to avoid key enumeration via distinct errors.
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: "invalid api key"})
	case errors.Is(err, domain.ErrMerchantSuspended):
		return c.JSON(http.StatusForbidden, errorResponse{Error: "merchant is suspended"})
	default:
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
