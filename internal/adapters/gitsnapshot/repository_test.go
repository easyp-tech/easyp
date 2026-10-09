package gitsnapshot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatAwareGitCommandsUseOperationTimeout(t *testing.T) {
	t.Setenv("EASYP_GIT_TIMEOUT", "50ms")
	backend := &repositoryFS{ctx: t.Context()}
	_, err := backend.command(nil, "-c", "alias.easyp-wait=!exec sleep 1", "easyp-wait")
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.NoError(t, t.Context().Err())
}
