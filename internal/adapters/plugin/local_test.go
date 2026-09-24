package plugin

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/easyp-tech/easyp/internal/adapters/console"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestLocalPluginExecutionFailure(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable uses a POSIX shell")
	}
	tests := []struct {
		name     string
		script   string
		canceled bool
		exitCode int
		stderr   string
	}{
		{name: "nonzero exit retains diagnostics", script: "echo plugin-failed >&2\nexit 7\n", exitCode: 7, stderr: "plugin-failed\n"},
		{name: "canceled context does not run plugin", script: "touch ran\n", canceled: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "test-plugin")
			require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"+tt.script), 0o755))
			ctx, cancel := context.WithCancel(t.Context())
			t.Cleanup(cancel)
			if tt.canceled {
				cancel()
			}
			executor := NewLocalPluginExecutor(logger.NewNop())

			response, err := executor.Execute(ctx, Info{Source: "./test-plugin", WorkDir: root}, &pluginpb.CodeGeneratorRequest{})

			require.Error(t, err)
			assert.Nil(t, response)
			var runErr *console.RunError
			require.ErrorAs(t, err, &runErr)
			assert.Equal(t, path, runErr.Command)
			assert.Equal(t, root, runErr.Dir)
			assert.Equal(t, tt.stderr, runErr.Stderr)
			if tt.canceled {
				assert.ErrorIs(t, runErr.Err, context.Canceled)
				assert.NoFileExists(t, filepath.Join(root, "ran"))
				return
			}
			var exitErr *exec.ExitError
			require.ErrorAs(t, runErr.Err, &exitErr)
			assert.Equal(t, tt.exitCode, exitErr.ExitCode())
		})
	}
}
