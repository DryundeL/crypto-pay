package invoice

import (
	"github.com/labstack/echo/v4"

	invoicehttp "github.com/DryundeL/crypto-pay/internal/modules/invoice/internal/delivery/http"
)

// RegisterRoutes is a thin facade required by Go internal-package visibility.
// Handlers and Echo usage live in internal/delivery/http.
func RegisterRoutes(g *echo.Group) {
	invoicehttp.RegisterRoutes(g)
}
