package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant"
	"github.com/DryundeL/crypto-pay/internal/modules/payment"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook"
	"github.com/DryundeL/crypto-pay/internal/modules/withdrawal"
	"github.com/DryundeL/crypto-pay/internal/platform/database"
	"github.com/DryundeL/crypto-pay/internal/platform/httpserver"
	"github.com/DryundeL/crypto-pay/internal/platform/observability"
)

const serviceName = "crypto-pay"

// App is the composition root for the API process.
type App struct {
	Config *Config
	DB     *gorm.DB
	Echo   *echo.Echo
	log    *slog.Logger
}

// Bootstrap wires platform + module HTTP delivery.
// Synchronous handlers are injected explicitly (no reflection bus).
func Bootstrap() (*App, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log := observability.Logger()

	db, err := database.Connect(cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if err := database.RunMigrationsUp(cfg.Database.DSN(), database.LogIdentity{
		User:   cfg.Database.User,
		Host:   cfg.Database.Host,
		Port:   cfg.Database.Port,
		DBName: cfg.Database.DBName,
	}, "migrations"); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	e := httpserver.New()
	httpserver.RegisterHealthRoutes(e, httpserver.NewHealthHandler(db, serviceName, cfg.App.Version))

	merchantModule := merchant.NewModule(merchant.Dependencies{
		DB:           db,
		APIKeyPepper: cfg.App.JWTSecret,
	})

	api := e.Group("/api")
	merchantModule.RegisterHTTP(api)
	invoice.RegisterRoutes(api)
	payment.RegisterRoutes(api)
	blockchain.RegisterRoutes(api)
	ledger.RegisterRoutes(api)
	withdrawal.RegisterRoutes(api)
	webhook.RegisterRoutes(api)

	return &App{
		Config: cfg,
		DB:     db,
		Echo:   e,
		log:    log,
	}, nil
}

// Run starts the HTTP server and blocks until ctx is cancelled.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		a.log.Info("server starting", "port", a.Config.App.Port, "env", a.Config.App.Env)
		if err := a.Echo.Start(":" + a.Config.App.Port); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.log.Info("application shutting down")
		if err := a.Echo.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http: %w", err)
		}
		return a.Close()
	case err := <-errCh:
		return err
	}
}

// Close releases process resources.
func (a *App) Close() error {
	if a.DB == nil {
		return nil
	}
	sqlDB, err := a.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close sql db: %w", err)
	}
	return nil
}
