package observability

import (
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = New(slog.LevelInfo)
}

// Logger returns the process-wide structured logger.
func Logger() *slog.Logger {
	return defaultLogger
}

// New creates a JSON slog logger at the given level.
func New(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
