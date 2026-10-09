package gitcommand

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultGitTimeoutIsFiveMinutes(t *testing.T) {
	t.Setenv("EASYP_GIT_TIMEOUT", "")
	timeout, err := Timeout()
	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, timeout)
}

func TestEarlierCallerDeadlineWins(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer cancel()
	_, err := Run(ctx, "", nil, "-c", "alias.easyp-wait=!exec sleep 1", "easyp-wait")
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
}

func TestCancelledCallerDoesNotRunGit(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := Run(ctx, "", nil, "--version")
	require.ErrorIs(t, err, context.Canceled)
}

func TestGitStdoutDoesNotContainStderr(t *testing.T) {
	t.Parallel()
	data, err := Run(t.Context(), "", nil, "-c", "alias.easyp-output=!printf committed; printf diagnostic >&2", "easyp-output")
	require.NoError(t, err)
	assert.Equal(t, "committed", string(data))
}
