package withdrawal

import (
	"github.com/labstack/echo/v4"

	withdrawalhttp "github.com/DryundeL/crypto-pay/internal/modules/withdrawal/internal/delivery/http"
)

func RegisterRoutes(g *echo.Group) {
	withdrawalhttp.RegisterRoutes(g)
}
