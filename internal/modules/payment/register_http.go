package payment

import (
	"github.com/labstack/echo/v4"

	paymenthttp "github.com/DryundeL/crypto-pay/internal/modules/payment/internal/delivery/http"
)

func RegisterRoutes(g *echo.Group) {
	paymenthttp.RegisterRoutes(g)
}
