package http

import "github.com/labstack/echo/v4"

type RouteDeps struct {
	Handler        *Handler
	AuthMiddleware echo.MiddlewareFunc
}

// RegisterRoutes mounts Payment HTTP API (merchant read-only).
func RegisterRoutes(g *echo.Group, deps RouteDeps) {
	payments := g.Group("/payments", deps.AuthMiddleware)
	payments.GET("/:id", deps.Handler.GetPayment)
}
