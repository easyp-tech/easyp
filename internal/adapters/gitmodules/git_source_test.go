package gitmodules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV1GitModuleCandidates(t *testing.T) {
	t.Parallel()
	localRoot := filepath.Join(t.TempDir(), "repo")
	local := filepath.Join(localRoot, "foo", "bar")
	require.NoError(t, os.MkdirAll(local, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(local, "protobuf.mod"), []byte("module "+filepath.ToSlash(local)+"\n"), 0o644))
	runTestGit(t, localRoot, "init", "-q")
	tests := []struct {
		name       string
		source     string
		wantRemote string
		wantDir    string
		wantTag    string
	}{
		{name: "bare module path", source: "github.com/acme/repo/foo/bar", wantRemote: "https://github.com/acme/repo", wantDir: "foo/bar", wantTag: "foo/bar/v1.2.3"},
		{name: "HTTPS module path", source: "https://example.com/acme/repo.git/foo", wantRemote: "https://example.com/acme/repo.git", wantDir: "foo", wantTag: "foo/v1.2.3"},
		{name: "local Git repository", source: local, wantRemote: localRoot, wantDir: filepath.Join("foo", "bar"), wantTag: "foo/bar/v1.2.3"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			candidates, err := v1GitModuleCandidates(tt.source)
			require.NoError(t, err)
			require.NotEmpty(t, candidates)
			var found bool
			for _, candidate := range candidates {
				if candidate.remote == tt.wantRemote && candidate.subdir == tt.wantDir && candidate.tag("v1.2.3") == tt.wantTag {
					found = true
				}
			}
			require.True(t, found)
		})
	}
}

func TestV1GitModuleCandidatesRejectTraversal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
	}{
		{name: "bare path", source: "github.com/acme/../repo"},
		{name: "HTTPS path", source: "https://example.com/acme/../repo"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := v1GitModuleCandidates(tt.source)

			require.ErrorContains(t, err, "invalid Git module")
		})
	}
}

func TestLocalV1GitModuleCandidatesDoNotInheritUnrelatedWorktree(t *testing.T) {
	t.Parallel()

	outer := t.TempDir()
	runTestGit(t, outer, "init", "-q")
	source := filepath.Join(outer, "tmp", "not-a-repository")
	require.NoError(t, os.MkdirAll(source, 0o755))

	candidates, err := v1GitModuleCandidates(source)

	require.NoError(t, err)
	require.Equal(t, []v1GitModuleCandidate{{remote: source}}, candidates)
}
