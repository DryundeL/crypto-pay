package merchant

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/command"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/query"
	merchanthttp "github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/delivery/http"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/domain"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/infrastructure/read"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/infrastructure/write"
	"github.com/DryundeL/crypto-pay/internal/platform/outbox"
	"github.com/DryundeL/crypto-pay/internal/platform/transaction"
)

// Module is the Merchant bounded context composition root (within the module).
type Module struct {
	handler     *merchanthttp.Handler
	hasher      *write.APIKeyGenerator
	auth        *read.APIKeyAuthenticator
	getMerchant *query.GetMerchantHandler
}

type Dependencies struct {
	DB           *gorm.DB
	APIKeyPepper string
}

func NewModule(deps Dependencies) *Module {
	tx := transaction.NewGormManager(deps.DB)
	outboxStore := outbox.NewStore(deps.DB)
	publisher := write.NewOutboxPublisher(outboxStore)
	repo := write.NewMerchantRepository(deps.DB)
	queries := read.NewMerchantQueries(deps.DB)
	keyGen := write.NewAPIKeyGenerator(deps.APIKeyPepper)
	authenticator := read.NewAPIKeyAuthenticator(deps.DB)

	createMerchant := command.NewCreateMerchantHandler(repo, tx, publisher, keyGen)
	createAPIKey := command.NewCreateAPIKeyHandler(repo, tx, publisher, keyGen)
	revokeAPIKey := command.NewRevokeAPIKeyHandler(repo, tx, publisher)
	getMerchant := query.NewGetMerchantHandler(queries)
	listAPIKeys := query.NewListAPIKeysHandler(queries)

	return &Module{
		handler: merchanthttp.NewHandler(
			createMerchant,
			createAPIKey,
			revokeAPIKey,
			getMerchant,
			listAPIKeys,
		),
		hasher:      keyGen,
		auth:        authenticator,
		getMerchant: getMerchant,
	}
}

func (m *Module) Name() string { return "merchant" }

func (m *Module) RegisterHTTP(g *echo.Group) {
	merchanthttp.RegisterRoutes(g, merchanthttp.RouteDeps{
		Handler: m.handler,
		Hasher:  m.hasher,
		Auth:    m.auth,
	})
}

// APIKeyMiddleware exposes merchant API-key auth for merchant-owned routes
// where path param "id" is the merchant id.
func (m *Module) APIKeyMiddleware() echo.MiddlewareFunc {
	return merchanthttp.APIKeyAuth(m.hasher, m.auth)
}

// APIKeyAuthOnly exposes API-key auth without binding path param "id" to merchant.
// Other modules (invoice, …) should use this.
func (m *Module) APIKeyAuthOnly() echo.MiddlewareFunc {
	return merchanthttp.APIKeyAuthOnly(m.hasher, m.auth)
}

// WebhookURL returns the merchant webhook endpoint URL (empty if not configured).
func (m *Module) WebhookURL(ctx context.Context, merchantID string) (string, error) {
	dto, err := m.getMerchant.Handle(ctx, query.GetMerchant{MerchantID: merchantID})
	if err != nil {
		if errors.Is(err, domain.ErrMerchantNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return dto.WebhookURL, nil
}
