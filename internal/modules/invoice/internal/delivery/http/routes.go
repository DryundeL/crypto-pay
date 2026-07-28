package http

import "github.com/labstack/echo/v4"

type RouteDeps struct {
	Handler        *Handler
	AuthMiddleware echo.MiddlewareFunc
}

// RegisterRoutes mounts Invoice HTTP API.
// Echo is confined to delivery/http.
func RegisterRoutes(g *echo.Group, deps RouteDeps) {
	invoices := g.Group("/invoices", deps.AuthMiddleware)
	invoices.POST("", deps.Handler.CreateInvoice)
	invoices.GET("/:id", deps.Handler.GetInvoice)
	invoices.POST("/:id/cancel", deps.Handler.CancelInvoice)
}
