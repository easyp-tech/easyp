package gitcommand

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/easyp-tech/easyp/internal/logger"
)

type loggerKey struct{}

// Debug adds phase details to the current dependency's diagnostics.
func Debug(ctx context.Context, message string, attrs ...slog.Attr) {
	if log, ok := ctx.Value(loggerKey{}).(logger.Logger); ok {
		log.Debug(ctx, message, attrs...)
	}
}

// WithLogger scopes dependency diagnostics without changing resolution inputs.
// Nested Git and snapshot operations inherit the module and revision attributes.
func WithLogger(ctx context.Context, log logger.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, log)
}

// WithAttributes adds diagnostic scope to an existing dependency logger.
func WithAttributes(ctx context.Context, attrs ...slog.Attr) context.Context {
	if log, ok := ctx.Value(loggerKey{}).(logger.Logger); ok {
		return WithLogger(ctx, log.With(attrs...))
	}
	return ctx
}

// Start reports a phase's start, periodic progress and final elapsed time.
// Call the returned function exactly once with the operation's result.
func Start(ctx context.Context, phase string, attrs ...slog.Attr) func(error) {
	log, ok := ctx.Value(loggerKey{}).(logger.Logger)
	if !ok {
		return func(error) {}
	}
	return startProgress(ctx, log.With(attrs...), phase, 15*time.Second)
}

func startProgress(ctx context.Context, log logger.Logger, phase string, interval time.Duration) func(error) {
	started := time.Now()
	log = log.With(slog.String("phase", phase))
	log.Debug(ctx, "Dependency operation started")
	stop := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Info(ctx, "Dependency operation still running", slog.Duration("elapsed", time.Since(started)))
			case <-ctx.Done():
				return
			case <-stop:
				return
			}
		}
	}()
	return func(err error) {
		close(stop)
		<-stopped
		status := "complete"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = "timeout"
		case errors.Is(err, context.Canceled):
			status = "canceled"
		case err != nil:
			status = "failed"
		}
		log.Debug(ctx, "Dependency operation finished", slog.Duration("elapsed", time.Since(started)), slog.String("status", status))
	}
}
