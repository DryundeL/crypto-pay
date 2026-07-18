package http

import "github.com/labstack/echo/v4"

// RegisterRoutes mounts Invoice HTTP API.
// Echo is confined to delivery/http.
func RegisterRoutes(g *echo.Group) {
	_ = g.Group("/invoices")
}
