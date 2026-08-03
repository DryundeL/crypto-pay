package http

import "github.com/labstack/echo/v4"

type RouteDeps struct {
	Handler        *Handler
	AuthMiddleware echo.MiddlewareFunc
}

// RegisterRoutes mounts Withdrawal HTTP API.
// Echo is confined to delivery/http.
func RegisterRoutes(g *echo.Group, deps RouteDeps) {
	withdrawals := g.Group("/withdrawals", deps.AuthMiddleware)
	withdrawals.POST("", deps.Handler.CreateWithdrawal)
	withdrawals.GET("", deps.Handler.ListWithdrawals)
	withdrawals.GET("/:id", deps.Handler.GetWithdrawal)
}
