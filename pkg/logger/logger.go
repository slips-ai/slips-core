package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// New creates a new structured logger with tint handler
func New(isDevelopment bool) *slog.Logger {
	var handler slog.Handler

	if isDevelopment {
		// Use tint for colorful development logs
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: "15:04:05",
		})
	} else {
		// Use JSON for production
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return slog.New(handler)
}
