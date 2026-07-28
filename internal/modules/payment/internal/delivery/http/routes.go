package http

import "github.com/labstack/echo/v4"

type RouteDeps struct {
	Handler        *Handler
	AuthMiddleware echo.MiddlewareFunc
}

// RegisterRoutes mounts Payment HTTP API.
func RegisterRoutes(g *echo.Group, deps RouteDeps) {
	payments := g.Group("/payments", deps.AuthMiddleware)
	payments.POST("/observed", deps.Handler.RecordObserved)
	payments.POST("/confirmed", deps.Handler.Confirm)
	payments.GET("/:id", deps.Handler.GetPayment)
}
