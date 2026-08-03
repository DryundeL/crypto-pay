package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DryundeL/crypto-pay/internal/app"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook"
	"github.com/DryundeL/crypto-pay/internal/platform/database"
	"github.com/DryundeL/crypto-pay/internal/platform/observability"
)

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
			result, err := webhookModule.ProcessDue(ctx, 50)
			if err != nil {
				log.Error("process due webhooks failed", "error", err)
				continue
			}
			if result.Processed > 0 {
				log.Info("processed webhook deliveries",
					"processed", result.Processed,
					"sent", result.Sent,
					"failed", result.Failed,
					"retried", result.Retried,
				)
			}
		}
	}
}
