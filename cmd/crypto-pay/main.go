package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/DryundeL/crypto-pay/internal/app"
)

func main() {
	application, err := app.Bootstrap()
	if err != nil {
		slog.Error("bootstrap failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		slog.Error("application stopped with error", "error", err)
		os.Exit(1)
	}

	slog.Info("application shut down successfully")
}
