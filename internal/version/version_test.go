package version

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMetadataFromBuildInfoReleaseVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		release     string
		mainVersion string
		revision    string
		modified    bool
		missingInfo bool
		wantVersion string
		wantNumbers [3]int32
		wantSuffix  string
	}{
		{
			name: "stable release overrides local VCS", release: "v1.0.0", mainVersion: develMainVersion,
			revision: "abcdef1234567890", modified: true,
			wantVersion: "v1.0.0", wantNumbers: [3]int32{1, 0, 0}, wantSuffix: "easyp",
		},
		{
			name: "nightly release overrides main module", release: "v1.0.0-nightly.20261005", mainVersion: "v0.17.1",
			revision: "abcdef1234567890", modified: true,
			wantVersion: "v1.0.0-nightly.20261005", wantNumbers: [3]int32{1, 0, 0}, wantSuffix: "easyp-nightly.20261005",
		},
		{
			name: "stable release without build info", release: "v1.0.0", missingInfo: true,
			wantVersion: "v1.0.0", wantNumbers: [3]int32{1, 0, 0}, wantSuffix: "easyp",
		},
		{
			name: "nightly release without build info", release: "v1.0.0-nightly.20261005", missingInfo: true,
			wantVersion: "v1.0.0-nightly.20261005", wantNumbers: [3]int32{1, 0, 0}, wantSuffix: "easyp-nightly.20261005",
		},
		{
			name: "empty override preserves go install version", mainVersion: "v1.2.3",
			wantVersion: "v1.2.3", wantNumbers: [3]int32{1, 2, 3}, wantSuffix: "easyp",
		},
		{
			name: "empty override preserves VCS revision", mainVersion: develMainVersion, revision: "abcdef1234567890",
			wantVersion: "abcdef1", wantSuffix: "easyp-abcdef1",
		},
		{
			name: "empty override preserves modified VCS revision", mainVersion: develMainVersion, revision: "abcdef1234567890", modified: true,
			wantVersion: "abcdef1-modified", wantSuffix: "easyp-abcdef1-modified",
		},
		{
			name: "empty override preserves main dirty marker normalization", mainVersion: "v1.2.3+dirty",
			wantVersion: "v1.2.3+modified", wantNumbers: [3]int32{1, 2, 3}, wantSuffix: "easyp-modified",
		},
		{
			name: "empty override without build info", missingInfo: true,
			wantVersion: develVersion, wantSuffix: "easyp-devel",
		},
		{
			name:        "empty override without module or VCS metadata",
			wantVersion: develVersion, wantSuffix: "easyp-devel",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var info *debug.BuildInfo
			if !tt.missingInfo {
				info = &debug.BuildInfo{Main: debug.Module{Version: tt.mainVersion}}
				if tt.revision != "" {
					info.Settings = append(info.Settings, debug.BuildSetting{Key: "vcs.revision", Value: tt.revision})
				}
				if tt.modified {
					info.Settings = append(info.Settings, debug.BuildSetting{Key: "vcs.modified", Value: "true"})
				}
			}

			metadata := buildMetadataFromBuildInfo(info, tt.release)

			assert.Equal(t, tt.wantVersion, metadata.easypVersion)
			require.NotNil(t, metadata.compilerVersion)
			assert.Equal(t, tt.wantSuffix, metadata.compilerVersion.GetSuffix())
			assert.Equal(t, tt.wantNumbers, [3]int32{
				metadata.compilerVersion.GetMajor(),
				metadata.compilerVersion.GetMinor(),
				metadata.compilerVersion.GetPatch(),
			})
		})
	}
}

func TestEasypVersionFromBuildInfo(t *testing.T) {
	t.Parallel()

	t.Run("stable main version", func(t *testing.T) {
		t.Parallel()

		bi := &debug.BuildInfo{
			Main: debug.Module{Version: "v1.2.3"},
		}

		if got, want := easypVersionFromBuildInfo(bi), "v1.2.3"; got != want {
			t.Fatalf("easypVersionFromBuildInfo() = %q, want %q", got, want)
		}
	})

	t.Run("maps dirty marker to modified in main version", func(t *testing.T) {
		t.Parallel()

		bi := &debug.BuildInfo{
			Main: debug.Module{Version: "v0.13.0-11-g6549e94-dirty"},
		}

		if got, want := easypVersionFromBuildInfo(bi), "v0.13.0-11-g6549e94-modified"; got != want {
			t.Fatalf("easypVersionFromBuildInfo() = %q, want %q", got, want)
		}
	})

	t.Run("devel main version with vcs fallback", func(t *testing.T) {
		t.Parallel()

		bi := &debug.BuildInfo{
			Main: debug.Module{Version: develMainVersion},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "1234567890abcdef"},
				{Key: "vcs.modified", Value: "true"},
				{Key: "vcs.time", Value: "2026-02-23T10:30:00Z"},
			},
		}

		if got, want := easypVersionFromBuildInfo(bi), "1234567-modified"; got != want {
			t.Fatalf("easypVersionFromBuildInfo() = %q, want %q", got, want)
		}
	})

	t.Run("devel main version with vcs revision", func(t *testing.T) {
		t.Parallel()

		bi := &debug.BuildInfo{
			Main: debug.Module{Version: develMainVersion},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "1234567890abcdef"},
				{Key: "vcs.modified", Value: "false"},
			},
		}

		if got, want := easypVersionFromBuildInfo(bi), "1234567"; got != want {
			t.Fatalf("easypVersionFromBuildInfo() = %q, want %q", got, want)
		}
	})

	t.Run("devel main version with no vcs metadata", func(t *testing.T) {
		t.Parallel()

		bi := &debug.BuildInfo{
			Main: debug.Module{Version: develMainVersion},
		}

		if got, want := easypVersionFromBuildInfo(bi), develVersion; got != want {
			t.Fatalf("easypVersionFromBuildInfo() = %q, want %q", got, want)
		}
	})

	t.Run("nil build info", func(t *testing.T) {
		t.Parallel()

		if got, want := easypVersionFromBuildInfo(nil), develVersion; got != want {
			t.Fatalf("easypVersionFromBuildInfo(nil) = %q, want %q", got, want)
		}
	})
}

func TestSemverMajorMinorPatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rawVersion string
		wantMajor  int32
		wantMinor  int32
		wantPatch  int32
	}{
		{
			name:       "stable semver",
			rawVersion: "v0.14.1",
			wantMajor:  0,
			wantMinor:  14,
			wantPatch:  1,
		},
		{
			name:       "without v prefix",
			rawVersion: "1.2.3",
			wantMajor:  1,
			wantMinor:  2,
			wantPatch:  3,
		},
		{
			name:       "pseudo version",
			rawVersion: "v0.0.0-20260223010101-abcdef123456",
			wantMajor:  0,
			wantMinor:  0,
			wantPatch:  0,
		},
		{
			name:       "invalid",
			rawVersion: "oops",
			wantMajor:  0,
			wantMinor:  0,
			wantPatch:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotMajor, gotMinor, gotPatch := semverMajorMinorPatch(tt.rawVersion)
			if gotMajor != tt.wantMajor || gotMinor != tt.wantMinor || gotPatch != tt.wantPatch {
				t.Fatalf("semverMajorMinorPatch(%q) = (%d,%d,%d), want (%d,%d,%d)",
					tt.rawVersion,
					gotMajor, gotMinor, gotPatch,
					tt.wantMajor, tt.wantMinor, tt.wantPatch,
				)
			}
		})
	}
}

func TestCompilerVersionFromEasyp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		easyp       string
		wantNumbers [3]int32
		wantSuffix  string
	}{
		{name: "stable", easyp: "v0.13.1", wantNumbers: [3]int32{0, 13, 1}, wantSuffix: "easyp"},
		{name: "modified", easyp: "v0.13.0-11-g6549e94-modified", wantNumbers: [3]int32{0, 13, 0}, wantSuffix: "easyp-11-g6549e94-modified"},
		{name: "devel", easyp: develVersion, wantSuffix: "easyp-devel"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			version := compilerVersionFromEasyp(tt.easyp)

			assert.Equal(t, tt.wantNumbers, [3]int32{version.GetMajor(), version.GetMinor(), version.GetPatch()})
			assert.Equal(t, tt.wantSuffix, version.GetSuffix())
		})
	}
}

func TestReplaceDirtyMarkers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "dash dirty", raw: "v0.13.0-11-g6549e94-dirty", want: "v0.13.0-11-g6549e94-modified"},
		{name: "build dirty", raw: "v0.13.0+dirty", want: "v0.13.0+modified"},
		{name: "dot dirty", raw: "v0.13.0+meta.dirty", want: "v0.13.0+meta.modified"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := replaceDirtyMarkers(tt.raw); got != tt.want {
				t.Fatalf("replaceDirtyMarkers(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestSanitizeSuffixToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "devel marker",
			raw:  develMainVersion,
			want: develVersion,
		},
		{
			name: "special chars",
			raw:  "v1.2.3+meta data",
			want: "v1.2.3_meta_data",
		},
		{
			name: "empty",
			raw:  "",
			want: unknownVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := sanitizeSuffixToken(tt.raw); got != tt.want {
				t.Fatalf("sanitizeSuffixToken(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
