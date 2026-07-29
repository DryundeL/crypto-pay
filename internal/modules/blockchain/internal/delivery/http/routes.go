package http

import "github.com/labstack/echo/v4"

type RouteDeps struct {
	Handler        *Handler
	AuthMiddleware echo.MiddlewareFunc
}

func RegisterRoutes(g *echo.Group, deps RouteDeps) {
	bc := g.Group("/blockchain", deps.AuthMiddleware)
	bc.POST("/observed", deps.Handler.RecordObservation)
	bc.POST("/confirmed", deps.Handler.ConfirmTransaction)
}
