package gitmodules

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV1GitModuleCandidates(t *testing.T) {
	t.Parallel()
	local := filepath.Join(t.TempDir(), "repo", "foo", "bar")
	tests := []struct {
		name       string
		source     string
		wantRemote string
		wantDir    string
		wantTag    string
	}{
		{name: "bare module path", source: "github.com/acme/repo/foo/bar", wantRemote: "https://github.com/acme/repo", wantDir: "foo/bar", wantTag: "foo/bar/v1.2.3"},
		{name: "HTTPS module path", source: "https://example.com/acme/repo.git/foo", wantRemote: "https://example.com/acme/repo.git", wantDir: "foo", wantTag: "foo/v1.2.3"},
		{name: "local Git repository", source: local, wantRemote: filepath.Dir(filepath.Dir(local)), wantDir: filepath.Join("foo", "bar"), wantTag: "foo/bar/v1.2.3"},
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
