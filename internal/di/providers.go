package di

import (

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/infrastructure/config"
	"github.com/DryundeL/crypto-pay/internal/infrastructure/database/pgsql"
)

// ConfigProvider provides application config
func ConfigProvider() (*config.Config, error) {
	return config.Load()
}

// DatabaseProvider provides GORM database connection
func DatabaseProvider(cfg *config.Config) (*gorm.DB, error) {
	return pgsql.Connect(cfg.Database.DSN())
}

// RouterParams contains deps for router creation
type RouterParams struct {
	fx.In

	DB      *gorm.DB
	Config  *config.Config
}

// NewRouter creates chi router with all routes
func NewRouter(p RouterParams) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)

	return router
}
