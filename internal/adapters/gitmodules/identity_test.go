package gitmodules

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceIdentityBoundsItsConfigurationCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("blocked Git fixture uses a POSIX shell")
	}
	t.Setenv("EASYP_GIT_TIMEOUT", "50ms")
	root := t.TempDir()
	bin := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bin, "git"), []byte("#!/bin/sh\ncase \"$1\" in\nrev-parse) pwd;;\n*) exec sleep 1;;\nesac\n"), 0o755))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	started := time.Now()
	assert.Empty(t, WorkspaceIdentity(t.Context(), root))
	assert.Less(t, time.Since(started), 800*time.Millisecond)
}

func TestWorkspaceIdentityUsesDeclaredOrigin(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, origin, want string }{
		{name: "HTTPS credentials", origin: "https://user:secret@example.test/acme/contracts.git", want: "example.test/acme/contracts"},
		{name: "SSH", origin: "git@example.test:acme/contracts.git", want: "example.test/acme/contracts"},
		{name: "local origin is not identity", origin: "/local/repository"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			git := func(args ...string) {
				cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", root}, args...)...)
				out, err := cmd.CombinedOutput()
				require.NoError(t, err, "%s", out)
			}
			git("init", "-q")
			git("remote", "add", "origin", tt.origin)
			git("config", "url.file:///local/mirror/.insteadOf", "https://user:secret@example.test/")
			assert.Equal(t, tt.want, WorkspaceIdentity(t.Context(), root))
			subdir := filepath.Join(root, "nested")
			require.NoError(t, os.Mkdir(subdir, 0o755))
			assert.Empty(t, WorkspaceIdentity(t.Context(), subdir), "a nested module must not borrow repository identity")
		})
	}
}
