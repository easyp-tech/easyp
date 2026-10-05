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

func TestLocalV1GitModuleCandidatesStopAtNearestWorktree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		exists bool
		file   string
	}{
		{name: "existing directory without live metadata", exists: true},
		{name: "directory removed in live checkout"},
		{name: "module directory replaced by a file", file: "foo/bar"},
		{name: "ancestor directory replaced by a file", file: "foo"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outer := t.TempDir()
			runTestGit(t, outer, "init", "-q")
			repository := filepath.Join(outer, "nested")
			require.NoError(t, os.MkdirAll(repository, 0o755))
			runTestGit(t, repository, "init", "-q")
			source := filepath.Join(repository, "foo", "bar")
			if tt.exists {
				require.NoError(t, os.MkdirAll(source, 0o755))
			}
			if tt.file != "" {
				path := filepath.Join(repository, tt.file)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
				require.NoError(t, os.WriteFile(path, []byte("no longer a directory\n"), 0o644))
			}

			candidates, err := v1GitModuleCandidates(source)

			require.NoError(t, err)
			require.Equal(t, []v1GitModuleCandidate{{remote: source}, {remote: repository, subdir: filepath.Join("foo", "bar")}}, candidates)
		})
	}
}
