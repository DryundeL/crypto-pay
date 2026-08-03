package http

import "github.com/labstack/echo/v4"

type RouteDeps struct {
	Handler        *Handler
	AuthMiddleware echo.MiddlewareFunc
}

// RegisterRoutes mounts Ledger HTTP API.
// Echo is confined to delivery/http.
func RegisterRoutes(g *echo.Group, deps RouteDeps) {
	ledger := g.Group("/ledger", deps.AuthMiddleware)
	ledger.GET("/balances", deps.Handler.ListBalances)
	ledger.GET("/journals", deps.Handler.ListJournals)
}
