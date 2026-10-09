package api

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/mod/sumdb/dirhash"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/migration"
	"github.com/easyp-tech/easyp/internal/modules"
)

type migrationBSRFixtureResolver struct {
	git        v1.Requirement
	resolution string
}

func (r migrationBSRFixtureResolver) Resolve(_ context.Context, dependency v1.BSRDependency) (v1.BSRResolution, error) {
	return v1.BSRResolution{Dependency: dependency, Git: r.git, Resolution: r.resolution}, nil
}

func TestMigrateBSRCompatibilityKeepsVerifiedHistoricalPins(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, resolution                   string
		nativeConflict, badHash, wantError bool
	}{
		{name: "approximate snapshot", resolution: v1.BSRCompatibilitySnapshot},
		{name: "exact BSR revision is not overridden", resolution: v1.BSRExactRevision, wantError: true},
		{name: "native Git constraint is not overridden", resolution: v1.BSRCompatibilitySnapshot, nativeConflict: true, wantError: true},
		{name: "historical integrity remains mandatory", resolution: v1.BSRCompatibilitySnapshot, badHash: true, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			provider, parent, consumer := t.TempDir(), t.TempDir(), t.TempDir()
			oldProto := "syntax = \"proto3\"; package google.api; message Annotation {}\n"
			writeV1GenerateFixture(t, provider, "google/api/annotations.proto", oldProto)
			historical := commitMigrationFixture(t, provider, true)
			writeV1GenerateFixture(t, provider, "google/api/annotations.proto", oldProto+"// a different compatibility snapshot\n")
			fallback := commitMigrationFixture(t, provider, false)
			config := "version: v2\nmodules: [{path: .}]\ndeps: [buf.build/googleapis/googleapis]\n"
			parentProto := "syntax = \"proto3\"; package gateway; message Gateway {}\n"
			writeV1GenerateFixture(t, parent, "buf.yaml", config)
			writeV1GenerateFixture(t, parent, "gateway.proto", parentProto)
			if tc.nativeConflict {
				writeV1GenerateFixture(t, parent, "protobuf.mod", "direct (\n"+provider+"@"+fallback+"\n)\n")
			}
			parentCommit := commitMigrationFixture(t, parent, true)
			providerHash := migrationFixtureProtoHash(t, "google/api/annotations.proto", oldProto)
			if tc.badHash {
				providerHash = "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
			}
			legacyLock := provider + " " + historical + " " + providerHash + "\n" + parent + " " + parentCommit + " " + migrationFixtureProtoHash(t, "gateway.proto", parentProto) + "\n"
			writeV1GenerateFixture(t, consumer, "easyp.yaml", "deps: ["+parent+", "+provider+"]\n")
			writeV1GenerateFixture(t, consumer, "easyp.lock", legacyLock)
			cache := gitmodules.NewWithBSRResolver(t.TempDir(), migrationBSRFixtureResolver{git: v1.Requirement{Module: provider, Version: fallback}, resolution: tc.resolution})
			plan, err := migration.Build(t.Context(), migration.Options{Dir: consumer, Module: "example.com/consumer", ResolveLock: true, Repository: cache})
			if tc.wantError {
				if tc.badHash {
					require.ErrorContains(t, err, "legacy hash mismatch")
				} else {
					require.ErrorContains(t, err, "incompatible requirement")
				}
				require.NoFileExists(t, filepath.Join(consumer, "protobuf.lock"))
				return
			}
			require.NoError(t, err)
			var lock v1.Lock
			for _, output := range plan.Outputs() {
				if output.Name == "protobuf.lock" {
					lock, err = v1.ParseLock(bytes.NewReader(output.Content))
					require.NoError(t, err)
				}
			}
			require.Len(t, lock.Modules, 2)
			for _, entry := range lock.Modules {
				if entry.Source == provider {
					require.Equal(t, historical, entry.Commit)
					require.Equal(t, historical, entry.Version)
				}
				if entry.Source == parent {
					require.Len(t, entry.BSR, 1)
					require.Equal(t, historical, entry.BSR[0].Git.Version)
					require.Equal(t, v1.BSRCompatibilitySnapshot, entry.BSR[0].Resolution)
				}
			}
			// Replay persisted bindings through the ordinary runtime cache. Merely
			// passing migration validation must not leave an unusable native lock.
			require.NoError(t, cache.Install(t.Context(), lock))
			_, err = modules.CachedSources(lock, cache)
			require.NoError(t, err)
			require.NoError(t, plan.Apply())
			raw, err := os.ReadFile(filepath.Join(consumer, "easyp.lock"))
			require.NoError(t, err)
			require.Equal(t, legacyLock, string(raw))
		})
	}
}

func commitMigrationFixture(t *testing.T, root string, initialize bool) string {
	t.Helper()
	if initialize {
		runTestGit(t, root, "init", "-q")
	}
	runTestGit(t, root, "add", ".")
	runTestGit(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "migration fixture")
	return strings.TrimSpace(runTestGit(t, root, "rev-parse", "HEAD"))
}

func migrationFixtureProtoHash(t *testing.T, name, content string) string {
	t.Helper()
	hash, err := dirhash.Hash1([]string{name}, func(string) (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(content)), nil })
	require.NoError(t, err)
	return hash
}
