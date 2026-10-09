// Package gitcommand runs bounded, non-interactive Git operations.
package gitcommand

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Timeout returns the positive per-operation limit configured by the user.
func Timeout() (time.Duration, error) {
	value := os.Getenv("EASYP_GIT_TIMEOUT")
	if value == "" {
		return 5 * time.Minute, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("ParseDuration: EASYP_GIT_TIMEOUT=%q: %w", value, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("EASYP_GIT_TIMEOUT must be a positive duration, got %q", value)
	}
	return duration, nil
}

// Operation bounds one command or cache-lock acquisition. Caller cancellation
// and an earlier caller deadline still take precedence.
func Operation(ctx context.Context) (context.Context, context.CancelFunc, error) {
	timeout, err := Timeout()
	if err != nil {
		return nil, nil, fmt.Errorf("Timeout: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	return ctx, cancel, nil
}

// Interrupted identifies cancellation and deadline failures even when the
// caller's own context remains live. Such failures must not trigger retries.
func Interrupted(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// Run returns stdout without mixing diagnostics into binary Git object data.
// Descendants and inherited pipes cannot keep a canceled command waiting forever.
func Run(ctx context.Context, directory string, stdin io.Reader, args ...string) (data []byte, err error) {
	ctx, cancel, err := Operation(ctx)
	if err != nil {
		return nil, fmt.Errorf("Operation: %w", err)
	}
	defer cancel()
	deadline, _ := ctx.Deadline()
	finish := Start(ctx, "git", slog.String("command", commandName(args)), slog.String("directory", directory), slog.Duration("timeout", time.Until(deadline)))
	defer func() { finish(err) }()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = directory
	cmd.Stdin = stdin
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.WaitDelay = 250 * time.Millisecond
	configureCancellation(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// A direct child can exit before its descendants close the inherited
		// pipes. WaitDelay may be hidden by an ExitError, so clean up on all
		// failures rather than relying on CommandContext's stopped watcher.
		if cmd.Process != nil {
			cancelErr := cmd.Cancel()
			if cancelErr != nil && !errors.Is(cancelErr, os.ErrProcessDone) {
				err = errors.Join(err, fmt.Errorf("Cancel: %w", cancelErr))
			}
		}
		if contextErr := ctx.Err(); contextErr != nil {
			return nil, fmt.Errorf("Run: git %s stopped (per-operation limit: EASYP_GIT_TIMEOUT): %w", commandName(args), contextErr)
		}
		if errors.Is(err, exec.ErrWaitDelay) {
			return nil, fmt.Errorf("Run: Git descendants kept output pipes open: %w", context.DeadlineExceeded)
		}
		return nil, fmt.Errorf("Run: git %s: %w: %s", commandName(args), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

func commandName(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "-c" || args[i] == "-C" {
			i++
			continue
		}
		return args[i]
	}
	return ""
}
