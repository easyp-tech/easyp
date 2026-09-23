package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV1GitModuleCandidates(t *testing.T) {
	t.Parallel()
	local := filepath.Join(t.TempDir(), "repo", "foo", "bar")
	for _, tc := range []struct {
		name       string
		source     string
		wantRemote string
		wantDir    string
		wantTag    string
	}{
		{name: "bare module path", source: "github.com/acme/repo/foo/bar", wantRemote: "https://github.com/acme/repo", wantDir: "foo/bar", wantTag: "foo/bar/v1.2.3"},
		{name: "HTTPS module path", source: "https://example.com/acme/repo.git/foo", wantRemote: "https://example.com/acme/repo.git", wantDir: "foo", wantTag: "foo/v1.2.3"},
		{name: "local Git repository", source: local, wantRemote: filepath.Dir(filepath.Dir(local)), wantDir: filepath.Join("foo", "bar"), wantTag: "foo/bar/v1.2.3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			candidates, err := v1GitModuleCandidates(tc.source)
			require.NoError(t, err)
			require.NotEmpty(t, candidates)
			var found bool
			for _, candidate := range candidates {
				if candidate.remote == tc.wantRemote && candidate.subdir == tc.wantDir && candidate.tag("v1.2.3") == tc.wantTag {
					found = true
				}
			}
			require.True(t, found)
		})
	}
}

func TestV1GitModuleCandidatesRejectTraversal(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"github.com/acme/../repo",
		"https://example.com/acme/../repo",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			_, err := v1GitModuleCandidates(source)
			require.Error(t, err)
		})
	}
}
