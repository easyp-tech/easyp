// Package logger provides a unified logging interface for the application.
package logger

import (
	"context"
	"log/slog"
)

// Logger is the application logging interface.
type Logger interface {
	Debug(ctx context.Context, msg string, attrs ...slog.Attr)
	Info(ctx context.Context, msg string, attrs ...slog.Attr)
	Warn(ctx context.Context, msg string, attrs ...slog.Attr)
	Error(ctx context.Context, msg string, attrs ...slog.Attr)
	With(attrs ...slog.Attr) Logger
}
