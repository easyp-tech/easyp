package logger

import "context"

type nopLogger struct{}

// NewNop creates a no-op Logger that discards all output. Useful for tests.
func NewNop() Logger { return &nopLogger{} }

func (n *nopLogger) Debug(context.Context, string, ...any) {}
func (n *nopLogger) Info(context.Context, string, ...any)  {}
func (n *nopLogger) Warn(context.Context, string, ...any)  {}
func (n *nopLogger) Error(context.Context, string, ...any) {}
func (n *nopLogger) With(...any) Logger                    { return n }
