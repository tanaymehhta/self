package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/tanaymehhta/self/backend/pkg/config"
)

type Logger struct {
	*slog.Logger
}

func New(cfg *config.Config) *Logger {
	var handler slog.Handler

	if cfg.IsDevelopment() {
		// Pretty text output for development
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
	} else {
		// JSON output for production
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: false,
		})
	}

	logger := slog.New(handler)
	return &Logger{Logger: logger}
}

func (l *Logger) WithRequest(requestID string, userID string) *Logger {
	return &Logger{
		Logger: l.With(
			"request_id", requestID,
			"user_id", userID,
			"timestamp", time.Now().UTC(),
		),
	}
}

func (l *Logger) WithUser(userID string) *Logger {
	return &Logger{
		Logger: l.With("user_id", userID),
	}
}

func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.With("component", component),
	}
}

func (l *Logger) LogError(err error, message string, args ...any) {
	l.Error(message, append([]any{"error", err.Error()}, args...)...)
}

func (l *Logger) LogHTTP(method, path string, status int, duration time.Duration, args ...any) {
	l.Info("HTTP Request",
		append([]any{
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
		}, args...)...,
	)
}