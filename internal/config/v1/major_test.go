package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMajorManifest(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, source, version string
		valid                 bool
	}{
		{"v0", "example.com/repo", "v0.9.0", true},
		{"v1", "example.com/repo", "v1.9.0", true},
		{"v2", "example.com/repo/v2", "v2.0.0", true},
		{"local", "/tmp/repo/v2", "v2.1.0", true},
		{"url", "ssh://git@example.com:2222/repo/v2", "v2.0.0", true},
		{"numeric_subdir", "example.com/repo/2024", "v1.0.0", true},
		{"v_word", "example.com/repo/v2beta", "v1.0.0", true},
		{"versionless", "example.com/repo/v2", "", true},
		{"sha", "example.com/repo/v2", strings.Repeat("a", 40), true},
		{"missing_suffix", "example.com/repo", "v2.0.0", false},
		{"wrong_major", "example.com/repo/v2", "v3.0.0", false},
		{"v0_suffix", "example.com/repo/v0", "", false},
		{"v1_suffix", "example.com/repo/v1", "", false},
		{"leading_zero", "example.com/repo/v01", "", false},
		{"decimal_suffix", "example.com/repo/v2.0", "", false},
		{"incompatible_v1", "example.com/repo", "v1.0.0+incompatible", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseModule(strings.NewReader("module example.com/app\nrequire " + tt.source + " " + tt.version + "\n"))
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
	for _, source := range []string{"example.com/repo/v0", "example.com/repo/v1", "example.com/repo/v01"} {
		_, err := ParseModule(strings.NewReader("module " + source + "\n"))
		require.Error(t, err)
		_, err = ParseModule(strings.NewReader("module example.com/app\nreplace " + source + " => ./local\n"))
		require.Error(t, err)
	}
}

func TestMajorLock(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, source, version string }{
		{"missing_suffix", "example.com/repo", "v2.0.0"},
		{"wrong_major", "example.com/repo/v2", "v1.0.0"},
		{"invalid_suffix", "example.com/repo/v01", strings.Repeat("a", 40)},
		{"incompatible_suffixed", "example.com/repo/v2", "v2.0.0+incompatible"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, (Lock{Version: 1, Modules: []LockedModule{{Source: tt.source, Version: tt.version, Commit: strings.Repeat("a", 40), Hash: "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}}).Validate())
		})
	}
}
