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

func (s *slogLogger) Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	s.l.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

func (s *slogLogger) Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	s.l.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

func (s *slogLogger) Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	s.l.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

func (s *slogLogger) Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	s.l.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

func (s *slogLogger) With(attrs ...slog.Attr) Logger {
	return &slogLogger{l: slog.New(s.l.Handler().WithAttrs(attrs))}
}
