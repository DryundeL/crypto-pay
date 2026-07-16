package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DryundeL/crypto-pay/internal/di"
	"github.com/DryundeL/crypto-pay/internal/infrastructure/database"
)

func main() {
	cfg, err := di.ConfigProvider()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := di.DatabaseProvider(cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrationsUp(cfg, "migrations"); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	router := di.NewRouter(di.RouterParams{
		DB:      db,
		Config:  cfg,
	})
	handler := router

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Server starting", "port", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Application shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	if sqlDB, err := db.DB(); err != nil {
		slog.Error("Failed to get database handle for close", "error", err)
	} else if err := sqlDB.Close(); err != nil {
		slog.Error("Failed to close database pool", "error", err)
	}

	slog.Info("Application shut down successfully")
}
