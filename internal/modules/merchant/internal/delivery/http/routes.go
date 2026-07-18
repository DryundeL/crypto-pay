package http

import (
	"github.com/labstack/echo/v4"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/ports"
)

type RouteDeps struct {
	Handler *Handler
	Hasher  ports.APIKeyGenerator
	Auth    ports.APIKeyAuthenticator
}

func RegisterRoutes(g *echo.Group, deps RouteDeps) {
	merchants := g.Group("/merchants")

	// Bootstrap: creates merchant + first API key (plaintext once).
	merchants.POST("", deps.Handler.CreateMerchant)

	secured := merchants.Group("", APIKeyAuth(deps.Hasher, deps.Auth))
	secured.GET("/:id", deps.Handler.GetMerchant)
	secured.POST("/:id/api-keys", deps.Handler.CreateAPIKey)
	secured.GET("/:id/api-keys", deps.Handler.ListAPIKeys)
	secured.DELETE("/:id/api-keys/:keyId", deps.Handler.RevokeAPIKey)
}
