package logger

import (
	"context"
	"log/slog"
)

type slogLogger struct {
	l *slog.Logger
}

// New creates a Logger backed by the given *slog.Logger.
func New(l *slog.Logger) Logger {
	return &slogLogger{l: l}
}

func (s *slogLogger) Debug(ctx context.Context, msg string, args ...any) {
	s.l.DebugContext(ctx, msg, args...)
}

func (s *slogLogger) Info(ctx context.Context, msg string, args ...any) {
	s.l.InfoContext(ctx, msg, args...)
}

func (s *slogLogger) Warn(ctx context.Context, msg string, args ...any) {
	s.l.WarnContext(ctx, msg, args...)
}

func (s *slogLogger) Error(ctx context.Context, msg string, args ...any) {
	s.l.ErrorContext(ctx, msg, args...)
}

func (s *slogLogger) With(args ...any) Logger {
	return &slogLogger{l: s.l.With(args...)}
}
