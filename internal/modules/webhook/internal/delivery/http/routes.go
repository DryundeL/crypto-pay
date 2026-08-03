package http

import "github.com/labstack/echo/v4"

type RouteDeps struct {
	Handler        *Handler
	AuthMiddleware echo.MiddlewareFunc
}

// RegisterRoutes mounts Webhook HTTP API.
// Echo is confined to delivery/http.
func RegisterRoutes(g *echo.Group, deps RouteDeps) {
	deliveries := g.Group("/webhook-deliveries", deps.AuthMiddleware)
	deliveries.GET("", deps.Handler.ListDeliveries)
	deliveries.GET("/:id", deps.Handler.GetDelivery)
}
