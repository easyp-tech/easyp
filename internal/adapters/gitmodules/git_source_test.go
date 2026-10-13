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
			candidates, err := v1GitModuleCandidates(tt.source, "")
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

			_, err := v1GitModuleCandidates(tt.source, "")

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

			candidates, err := v1GitModuleCandidates(source, "")

			require.NoError(t, err)
			require.Equal(t, []v1GitModuleCandidate{{remote: source}, {remote: repository, subdir: filepath.Join("foo", "bar")}}, candidates)
		})
	}
}

func TestV1GitModuleCandidatesProxy(t *testing.T) {
	t.Parallel()

	localRoot := filepath.Join(t.TempDir(), "repo")
	require.NoError(t, os.MkdirAll(localRoot, 0o755))
	runTestGit(t, localRoot, "init", "-q")

	tests := []struct {
		name    string
		source  string
		proxy   string
		remotes []string
	}{
		{
			name:    "path_host_replaced",
			source:  "github.com/group/package",
			proxy:   "gitlab.com",
			remotes: []string{"https://gitlab.com/group/package", "https://gitlab.com/group"},
		},
		{
			name:    "path_host_and_port_replaced",
			source:  "github.com/group/package",
			proxy:   "gitlab.example.com:8443",
			remotes: []string{"https://gitlab.example.com:8443/group/package", "https://gitlab.example.com:8443/group"},
		},
		{
			name:    "empty_proxy_keeps_host",
			source:  "github.com/group/package",
			remotes: []string{"https://github.com/group/package", "https://github.com/group"},
		},
		{
			name:    "https_host_replaced",
			source:  "https://example.com/acme/repo.git/foo",
			proxy:   "gitlab.com",
			remotes: []string{"https://gitlab.com/acme/repo.git/foo", "https://gitlab.com/acme/repo.git", "https://gitlab.com/acme"},
		},
		{
			name:    "ssh_host_replaced",
			source:  "ssh://git@example.com:2222/acme/repo",
			proxy:   "gitlab.example.com:8443",
			remotes: []string{"ssh://git@gitlab.example.com:8443/acme/repo", "ssh://git@gitlab.example.com:8443/acme"},
		},
		{
			name:    "local_path_unchanged",
			source:  localRoot,
			proxy:   "gitlab.com",
			remotes: []string{localRoot},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			candidates, err := v1GitModuleCandidates(tt.source, tt.proxy)

			require.NoError(t, err)
			remotes := make([]string, 0, len(candidates))
			for _, candidate := range candidates {
				remotes = append(remotes, candidate.remote)
			}
			require.Equal(t, tt.remotes, remotes)
		})
	}
}
