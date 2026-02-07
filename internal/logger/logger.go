// Package logger provides a unified logging interface for the application.
package logger

import "context"

// Logger is the application logging interface.
type Logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
}
