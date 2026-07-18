package main

import (
	"log/slog"
	"os"
)

// Worker process entrypoint (Webhook / async jobs).
// Will process outbox events, webhook deliveries, invoice expiration, etc.
func main() {
	slog.Info("worker stub — not implemented yet")
	os.Exit(0)
}
