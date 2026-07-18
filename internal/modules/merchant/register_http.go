package merchant

import "github.com/labstack/echo/v4"

// RegisterRoutes mounts Merchant HTTP API via the wired module.
func RegisterRoutes(g *echo.Group, m *Module) {
	m.RegisterHTTP(g)
}
