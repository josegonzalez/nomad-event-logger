package agent

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

// initLogger initializes the JSON logger
func initLogger() {
	// Create JSON handler with default options
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger = slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)
}

// GetLogger returns the configured logger
func GetLogger() *slog.Logger {
	if logger == nil {
		initLogger()
	}
	return logger
}
