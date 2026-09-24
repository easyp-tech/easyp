package gitmodules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchV1ModuleRemovesTemporaryCheckouts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		version    string
		useCommit  bool
		invalidMod bool
		wantErr    string
	}{
		{name: "head"},
		{name: "commit", useCommit: true},
		{name: "tag", version: "v1.0.0"},
		{name: "missing tag", version: "v1.1.0", wantErr: "was not found"},
		{name: "invalid module at head", invalidMod: true, wantErr: "declares module"},
		{name: "invalid module at commit", invalidMod: true, useCommit: true, wantErr: "declares module"},
		{name: "invalid module at tag", invalidMod: true, version: "v1.0.0", wantErr: "declares module"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			remote := t.TempDir()
			moduleName := remote
			if tt.invalidMod {
				moduleName = "example.com/other"
			}
			require.NoError(t, os.WriteFile(filepath.Join(remote, "protobuf.mod"), []byte("module "+moduleName+"\n"), 0o644))
			runTestGit(t, remote, "init", "-q")
			runTestGit(t, remote, "add", ".")
			runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
			runTestGit(t, remote, "tag", "v1.0.0")
			commit := strings.TrimSpace(runTestGit(t, remote, "rev-parse", "HEAD"))
			version := tt.version
			if tt.useCommit {
				version = commit
			}
			cache := t.TempDir()

			module, err := (&Cache{root: cache}).Fetch(t.Context(), remote, version)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, commit, module.Lock.Commit)
				assert.Equal(t, remote, module.Module.Name)
				assert.Equal(t, remote, module.Lock.Source)
			}
			entries, err := os.ReadDir(cache)
			require.NoError(t, err)
			assert.Empty(t, entries)
		})
	}
}
