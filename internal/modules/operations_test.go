package modules

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestTidyPreservesFilesOnDependencyFailure(t *testing.T) {
	t.Parallel()
	dependencyErr := errors.New("dependency failed")
	tests := []struct {
		name                            string
		fetchErr, installErr, cachedErr error
		wantInstall                     bool
	}{
		{name: "metadata fetch failed", fetchErr: dependencyErr},
		{name: "installation failed", installErr: dependencyErr, wantInstall: true},
		{name: "cached metadata failed", cachedErr: dependencyErr, wantInstall: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			original := []byte("module example.com/app\nrequire example.com/dep v1.0.0 // keep\n")
			require.NoError(t, os.WriteFile(filepath.Join(root, v1.ModuleFile), original, 0o600))
			lock := []byte("version: 1\nmodules: []\n")
			require.NoError(t, os.WriteFile(filepath.Join(root, v1.LockFile), lock, 0o600))
			repository := &fakeRepository{fetchErr: tt.fetchErr, installErr: tt.installErr, cachedErr: tt.cachedErr}

			err := Tidy(t.Context(), root, repository)

			require.ErrorIs(t, err, dependencyErr)
			assert.Equal(t, tt.wantInstall, repository.installed)
			manifestAfter, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
			require.NoError(t, err)
			assert.Equal(t, original, manifestAfter)
			lockAfter, err := os.ReadFile(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Equal(t, lock, lockAfter)
		})
	}
}

func TestAugmentManifestClassifiesTransitiveImports(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		suffix string
	}{
		{name: "imported transitive dependency is direct", source: "syntax = \"proto3\"; import \"dep.proto\";", suffix: "\n"},
		{name: "unreferenced transitive dependency is indirect", source: "syntax = \"proto3\";", suffix: " // indirect\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, dependency := t.TempDir(), t.TempDir()
			writeV1GenerateFixture(t, root, "root.proto", tt.source)
			writeV1GenerateFixture(t, dependency, "proto/dep.proto", "syntax = \"proto3\";")
			original := []byte("module example.com/app // keep\n")
			module := v1.Module{Name: "example.com/app", Roots: []string{"."}}
			lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{{Source: "example.com/dep", Version: "v1.0.0"}}}
			repository := &fakeRepository{directory: dependency, module: v1.Module{Name: "example.com/dep", Roots: []string{"proto"}}}

			updated, err := augmentV1ManifestRequirements(original, root, module, lock, repository)

			require.NoError(t, err)
			assert.Equal(t, "module example.com/app // keep\nrequire example.com/dep v1.0.0"+tt.suffix, string(updated))
			assert.Equal(t, "module example.com/app // keep\n", string(original))
			assert.False(t, repository.installed)
		})
	}
}

func TestLatestCompatibleTag(t *testing.T) {
	t.Parallel()
	const commit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	versionsErr := errors.New("cannot list versions")
	tests := []struct {
		name, current string
		versions      []string
		sourceErr     error
		want          string
		wantErr       error
		wantMessage   string
		wantRead      bool
	}{
		{name: "versionless uses HEAD without listing tags"},
		{name: "explicit commit stays pinned", current: commit, want: commit},
		{name: "stable stays in major and excludes prereleases", current: "v1.2.0", versions: []string{"v2.0.0", "v1.3.0-rc.1", "v1.2.1", "v1.1.0"}, want: "v1.2.1", wantRead: true},
		{name: "prerelease accepts newer prerelease", current: "v1.2.0-rc.1", versions: []string{"v1.2.0-rc.2", "v1.1.0"}, want: "v1.2.0-rc.2", wantRead: true},
		{name: "older tags never downgrade", current: "v1.2.0", versions: []string{"v1.1.0"}, want: "v1.2.0", wantRead: true},
		{name: "missing tags preserve current version", current: "v1.2.0", want: "v1.2.0", wantRead: true},
		{name: "invalid current version", current: "branch", wantMessage: "invalid required version"},
		{name: "listing error preserves cause", current: "v1.2.0", sourceErr: versionsErr, wantErr: versionsErr, wantRead: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := &fakeRepository{versions: tt.versions, versionsErr: tt.sourceErr}

			version, err := latestCompatibleV1Tag(t.Context(), "example.com/dep", tt.current, repository)

			assert.Equal(t, tt.wantRead, repository.readVersions)
			if tt.wantMessage != "" {
				require.ErrorContains(t, err, tt.wantMessage)
				return
			}
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, version)
		})
	}
}

func TestDownloadRejectsStaleLockBeforeInstallation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		lock string
	}{
		{name: "missing requirement", lock: "version: 1\nmodules: []\n"},
		{name: "older version", lock: "version: 1\nmodules:\n  - source: example.com/dep\n    version: v1.0.0\n    commit: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n    hash: h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			manifest := "module example.com/app\nrequire example.com/dep v1.1.0\n"
			writeV1GenerateFixture(t, root, v1.ModuleFile, manifest)
			writeV1GenerateFixture(t, root, v1.LockFile, tt.lock)
			cache := &fakeRepository{}

			err := Download(t.Context(), root, cache)

			require.ErrorContains(t, err, "protobuf.lock does not satisfy example.com/dep v1.1.0")
			assert.False(t, cache.installed)
			actualManifest, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
			require.NoError(t, err)
			assert.Equal(t, manifest, string(actualManifest))
			actualLock, err := os.ReadFile(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Equal(t, tt.lock, string(actualLock))
		})
	}
}

type fakeRepository struct {
	fetchErr, installErr, cachedErr, versionsErr error
	directory                                    string
	module                                       v1.Module
	versions                                     []string
	installed, readVersions                      bool
}

func (r *fakeRepository) Fetch(_ context.Context, source, version string) (Fetched, error) {
	return Fetched{Module: v1.Module{Name: source, Roots: []string{"."}}, Lock: v1.LockedModule{Source: source, Version: version, Commit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Hash: "h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="}}, r.fetchErr
}
func (r *fakeRepository) Install(context.Context, v1.Lock) error {
	r.installed = true
	return r.installErr
}
func (r *fakeRepository) Cached(v1.LockedModule) (string, v1.Module, error) {
	return r.directory, r.module, r.cachedErr
}
func (r *fakeRepository) Versions(context.Context, string) ([]string, error) {
	r.readVersions = true
	return r.versions, r.versionsErr
}
