package http

import "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group) {
	_ = g.Group("/ledger")
}
