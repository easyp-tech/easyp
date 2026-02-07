package logger

import (
	"context"
	"log/slog"
)

type nopLogger struct{}

// NewNop creates a no-op Logger that discards all output. Useful for tests.
func NewNop() Logger { return &nopLogger{} }

func (n *nopLogger) Debug(context.Context, string, ...slog.Attr) {}
func (n *nopLogger) Info(context.Context, string, ...slog.Attr)  {}
func (n *nopLogger) Warn(context.Context, string, ...slog.Attr)  {}
func (n *nopLogger) Error(context.Context, string, ...slog.Attr) {}
func (n *nopLogger) With(...slog.Attr) Logger                    { return n }
