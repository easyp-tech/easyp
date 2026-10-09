package gitcommand

import (
	"bytes"
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestProgressIdentifiesPhaseAndStopsAfterCompletion(t *testing.T) {
	t.Parallel()
	var output synchronizedBuffer
	log := logger.New(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	finish := startProgress(t.Context(), log.With(slog.String("source", "example.test/contracts")), "legacy archive proof", 5*time.Millisecond)
	require.Eventually(t, func() bool { return bytes.Contains(output.bytes(), []byte("still running")) }, time.Second, time.Millisecond)
	finish(context.DeadlineExceeded)
	data := output.bytes()
	assert.Contains(t, string(data), `"source":"example.test/contracts"`)
	assert.Contains(t, string(data), `"phase":"legacy archive proof"`)
	assert.Contains(t, string(data), `"elapsed":`)
	assert.Contains(t, string(data), `"status":"timeout"`)
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, data, output.bytes(), "progress must stop when the operation returns")
}

func TestGitCommandDiagnosticsPreserveStdout(t *testing.T) {
	t.Parallel()
	var output synchronizedBuffer
	log := logger.New(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	ctx := WithLogger(t.Context(), log)
	data, err := Run(ctx, "", nil, "--version")
	require.NoError(t, err)
	assert.Contains(t, string(data), "git version")
	assert.Contains(t, string(output.bytes()), `"phase":"git"`)
	assert.Contains(t, string(output.bytes()), `"status":"complete"`)
	assert.Contains(t, string(output.bytes()), `"timeout":`)
}

type synchronizedBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (b *synchronizedBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Write(data)
}

func (b *synchronizedBuffer) bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return bytes.Clone(b.buffer.Bytes())
}
