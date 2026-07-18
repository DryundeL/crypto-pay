package blockchain

import (
	"github.com/labstack/echo/v4"

	blockchainhttp "github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/delivery/http"
)

func RegisterRoutes(g *echo.Group) {
	blockchainhttp.RegisterRoutes(g)
}
