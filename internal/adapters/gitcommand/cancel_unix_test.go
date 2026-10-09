//go:build unix

package gitcommand

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInheritedOutputPipeDoesNotLeaveGitDescendantRunning(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		exit string
	}{
		{name: "parent succeeds", exit: "0"},
		{name: "parent fails", exit: "1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			_, err := Run(t.Context(), directory, nil, "-c", "alias.easyp-child=!sleep 30 & printf '%s' \"$!\" > child.pid; exit "+tt.exit, "easyp-child")
			if tt.exit == "0" {
				require.ErrorIs(t, err, context.DeadlineExceeded)
			} else {
				require.Error(t, err)
			}
			data, err := os.ReadFile(filepath.Join(directory, "child.pid"))
			require.NoError(t, err)
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			require.NoError(t, err)
			t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
			require.Eventually(t, func() bool {
				return errors.Is(syscall.Kill(pid, 0), syscall.ESRCH)
			}, time.Second, 10*time.Millisecond, "the inherited-pipe timeout must terminate the remaining Git group")
		})
	}
}
