package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DryundeL/crypto-pay/internal/app"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook"
	"github.com/DryundeL/crypto-pay/internal/platform/database"
	"github.com/DryundeL/crypto-pay/internal/platform/observability"
)

// workerAddressProvider satisfies invoice.Dependencies; create is unused in worker.
type workerAddressProvider struct{}

func (workerAddressProvider) AllocateAddress(context.Context, string, string, string) (string, error) {
	return "", fmt.Errorf("address allocation is not available in worker")
}

func main() {
	log := observability.Logger()

	cfg, err := app.LoadConfig()
	if err != nil {
		log.Error("load config failed", "error", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.Database.DSN())
	if err != nil {
		log.Error("connect database failed", "error", err)
		os.Exit(1)
	}

	invoiceModule := invoice.NewModule(invoice.Dependencies{
		DB:              db,
		AddressProvider: workerAddressProvider{},
	})
	merchantModule := merchant.NewModule(merchant.Dependencies{
		DB:           db,
		APIKeyPepper: cfg.App.JWTSecret,
	})
	webhookModule := webhook.NewModule(webhook.Dependencies{
		DB:            db,
		Merchant:      merchantModule,
		SigningSecret: cfg.App.JWTSecret,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("worker starting", "env", cfg.App.Env)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("worker shutting down")
			sqlDB, err := db.DB()
			if err == nil {
				_ = sqlDB.Close()
			}
			return
		case <-ticker.C:
			if err := runOnce(ctx, invoiceModule, webhookModule); err != nil {
				log.Error("worker tick failed", "error", err)
			}
		}
	}
}

func runOnce(ctx context.Context, invoices *invoice.Module, webhooks *webhook.Module) error {
	log := observability.Logger()

	expired, err := invoices.ExpireDue(ctx, 100)
	if err != nil {
		return fmt.Errorf("expire due invoices: %w", err)
	}
	if expired.Expired > 0 || expired.Skipped > 0 {
		log.Info("expired invoices",
			"expired", expired.Expired,
			"skipped", expired.Skipped,
		)
	}

	result, err := webhooks.ProcessDue(ctx, 50)
	if err != nil {
		return fmt.Errorf("process due webhooks: %w", err)
	}
	if result.Processed > 0 {
		log.Info("processed webhook deliveries",
			"processed", result.Processed,
			"sent", result.Sent,
			"failed", result.Failed,
			"retried", result.Retried,
		)
	}
	return nil
}
