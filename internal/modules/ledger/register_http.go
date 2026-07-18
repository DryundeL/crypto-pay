package ledger

import (
	"github.com/labstack/echo/v4"

	ledgerhttp "github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/delivery/http"
)

func RegisterRoutes(g *echo.Group) {
	ledgerhttp.RegisterRoutes(g)
}
