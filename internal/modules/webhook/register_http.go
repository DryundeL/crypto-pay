package webhook

import (
	"github.com/labstack/echo/v4"

	webhookhttp "github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/delivery/http"
)

func RegisterRoutes(g *echo.Group) {
	webhookhttp.RegisterRoutes(g)
}
