package migration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

const testCommit = "0123456789abcdef0123456789abcdef01234567"
const testHash = "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
const testNewHash = "h1:AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE="

type mockRepository struct {
	fetched       map[string]modules.Fetched
	calls         []string
	wantOldHash   string
	wantOldHashes map[string]string
	err           error
}

func (r *mockRepository) FetchMigration(_ context.Context, source, version, oldHash string) (modules.Fetched, error) {
	r.calls = append(r.calls, source+"@"+version)
	if r.err != nil {
		return modules.Fetched{}, r.err
	}
	wantOldHash := r.wantOldHash
	if r.wantOldHashes != nil {
		var ok bool
		wantOldHash, ok = r.wantOldHashes[source+"@"+version]
		if !ok {
			return modules.Fetched{}, fmt.Errorf("unexpected integrity request %s@%s", source, version)
		}
	}
	if oldHash != wantOldHash {
		return modules.Fetched{}, fmt.Errorf("unexpected legacy integrity input %s", oldHash)
	}
	value, ok := r.fetched[source+"@"+version]
	if !ok {
		return modules.Fetched{}, fmt.Errorf("unexpected revision %s@%s", source, version)
	}
	return value, nil
}

func TestMigrateLockTaggedHistoricalPins(t *testing.T) {
	t.Parallel()
	const dependency = "example.com/acme/dep"
	const consumer = "example.com/acme/consumer"
	for _, tt := range []struct {
		name, legacyVersion, directVersion, transitiveVersion, wantVersion string
	}{
		{name: "sha_direct", legacyVersion: testCommit, directVersion: "v1.2.3", wantVersion: "v1.2.3"},
		{name: "pseudo_direct", legacyVersion: "v0.0.0-20260101123456-" + testCommit, directVersion: "v1.2.3", wantVersion: "v1.2.3"},
		{name: "transitive_higher", legacyVersion: testCommit, directVersion: "v1.0.0", transitiveVersion: "v2.0.0", wantVersion: "v2.0.0"},
		{name: "direct_higher", legacyVersion: testCommit, directVersion: "v2.0.0", transitiveVersion: "v1.0.0", wantVersion: "v2.0.0"},
		{name: "tag_pin_retained", legacyVersion: "v3.0.0", directVersion: "v1.0.0", wantVersion: "v3.0.0"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pins, err := parseLegacyLock([]byte(dependency + " " + tt.legacyVersion + " " + testHash + "\n"))
			require.NoError(t, err)
			module := v1.Module{Name: "example.com/acme/local", Requires: []v1.Requirement{{Module: dependency, Version: tt.directVersion}}}
			pin := migrationFetched(dependency, pins[dependency].version, testCommit)
			repo := &mockRepository{
				fetched:       map[string]modules.Fetched{dependency + "@" + pin.Lock.Version: pin},
				wantOldHashes: map[string]string{dependency + "@" + pin.Lock.Version: testHash},
			}
			if tt.transitiveVersion != "" {
				module.Requires = append(module.Requires, v1.Requirement{Module: consumer})
				pins[consumer] = legacyPin{source: consumer, version: testCommit, hash: testHash}
				repo.fetched[consumer+"@"+testCommit] = migrationFetched(consumer, testCommit, testCommit, v1.Requirement{Module: dependency, Version: tt.transitiveVersion})
				repo.wantOldHashes[consumer+"@"+testCommit] = testHash
			}
			if pin.Lock.Version == testCommit {
				for _, version := range []string{tt.directVersion, tt.transitiveVersion} {
					if version != "" {
						repo.fetched[dependency+"@"+version] = migrationFetched(dependency, version, testCommit)
						repo.wantOldHashes[dependency+"@"+version] = ""
					}
				}
			}
			lock, err := migrateLock(t.Context(), module, pins, true, repo)
			require.NoError(t, err)
			require.NoError(t, modules.ValidateRequirements(module.Requires, lock))
			for _, entry := range lock.Modules {
				assert.Equal(t, testCommit, entry.Commit)
				assert.Equal(t, testNewHash, entry.Hash)
				if entry.Source == dependency {
					assert.Equal(t, tt.wantVersion, entry.Version)
				}
				metadata := repo.fetched[entry.Source+"@"+pins[entry.Source].version].Module
				require.NoError(t, modules.ValidateRequirements(metadata.Requires, lock))
			}
		})
	}
}

func TestMigrateLockDoesNotInventTagLabels(t *testing.T) {
	t.Parallel()
	const dependency = "example.com/acme/dep"
	const otherCommit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	for _, tt := range []struct {
		name, tagCommit, returnedVersion, tagHash, wantError string
		missingTag                                           bool
	}{
		{name: "tag_moved", tagCommit: otherCommit, returnedVersion: "v1.0.0", tagHash: testNewHash, wantError: "incompatible"},
		{name: "tag_missing", missingTag: true, wantError: "unexpected integrity request"},
		{name: "tag_label_differs", tagCommit: testCommit, returnedVersion: "v2.0.0", tagHash: testNewHash, wantError: "resolved version"},
		{name: "tag_hash_differs", tagCommit: testCommit, returnedVersion: "v1.0.0", tagHash: testHash, wantError: "content hash"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			module := v1.Module{Name: "example.com/acme/local", Requires: []v1.Requirement{{Module: dependency, Version: "v1.0.0"}}}
			pins := map[string]legacyPin{dependency: {source: dependency, version: testCommit, hash: testHash}}
			repo := &mockRepository{
				fetched:       map[string]modules.Fetched{dependency + "@" + testCommit: migrationFetched(dependency, testCommit, testCommit)},
				wantOldHashes: map[string]string{dependency + "@" + testCommit: testHash},
			}
			if !tt.missingTag {
				fetched := migrationFetched(dependency, tt.returnedVersion, tt.tagCommit)
				fetched.Lock.Hash = tt.tagHash
				repo.fetched[dependency+"@v1.0.0"] = fetched
				repo.wantOldHashes[dependency+"@v1.0.0"] = ""
			}
			lock, err := migrateLock(t.Context(), module, pins, true, repo)
			require.ErrorContains(t, err, tt.wantError)
			assert.Empty(t, lock.Modules)
		})
	}
}

func TestMigrateLockInitialResolution(t *testing.T) {
	t.Parallel()
	const dependency = "example.com/acme/dep"
	const transitive = "example.com/acme/transitive"
	for _, tt := range []struct {
		name, version, fetchedVersion, wantError string
	}{
		{name: "tagged_transitive", version: "v1.2.3", fetchedVersion: "v1.2.3"},
		{name: "pseudo_transitive_requires_manual_migration", version: "v0.0.0-20260101123456-" + testCommit, fetchedVersion: testCommit, wantError: "manual migration"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			module := v1.Module{Name: "example.com/acme/local", Requires: []v1.Requirement{{Module: dependency}}}
			dep := migrationFetched(dependency, testCommit, testCommit, v1.Requirement{Module: transitive, Version: tt.version})
			repo := &mockRepository{
				fetched: map[string]modules.Fetched{
					dependency + "@":                     dep,
					transitive + "@" + tt.fetchedVersion: migrationFetched(transitive, tt.fetchedVersion, testCommit),
				},
				wantOldHashes: map[string]string{dependency + "@": "", transitive + "@" + tt.fetchedVersion: ""},
			}
			lock, err := migrateLock(t.Context(), module, nil, false, repo)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				assert.Empty(t, lock.Modules)
				return
			}
			require.NoError(t, err)
			require.Len(t, lock.Modules, 2)
			require.NoError(t, modules.ValidateRequirements(module.Requires, lock))
			require.NoError(t, modules.ValidateRequirements(dep.Module.Requires, lock))
		})
	}
}

func TestMigrateLockRejectsHistoricalPseudoMetadataMismatch(t *testing.T) {
	t.Parallel()
	const dependency = "example.com/acme/dep"
	const transitive = "example.com/acme/transitive"
	module := v1.Module{Name: "example.com/acme/local", Requires: []v1.Requirement{{Module: dependency}}}
	pins := map[string]legacyPin{
		dependency: {source: dependency, version: testCommit, hash: testHash},
		transitive: {source: transitive, version: testCommit, hash: testHash},
	}
	repo := &mockRepository{
		fetched: map[string]modules.Fetched{
			dependency + "@" + testCommit: migrationFetched(dependency, testCommit, testCommit, v1.Requirement{Module: transitive, Version: "v0.0.0-20260101123456-" + testCommit}),
			transitive + "@" + testCommit: migrationFetched(transitive, testCommit, testCommit),
		},
		wantOldHash: testHash,
	}
	lock, err := migrateLock(t.Context(), module, pins, true, repo)
	require.ErrorContains(t, err, "manual migration")
	assert.Empty(t, lock.Modules)
}

func TestMigrateLockChecksEveryRequiredTagBeforeLabeling(t *testing.T) {
	t.Parallel()
	const dependency = "example.com/acme/dep"
	const consumer = "example.com/acme/consumer"
	module := v1.Module{Name: "example.com/acme/local", Requires: []v1.Requirement{{Module: dependency, Version: "v2.0.0"}, {Module: consumer}}}
	pins := map[string]legacyPin{
		dependency: {source: dependency, version: testCommit, hash: testHash},
		consumer:   {source: consumer, version: testCommit, hash: testHash},
	}
	repo := &mockRepository{
		fetched: map[string]modules.Fetched{
			dependency + "@" + testCommit: migrationFetched(dependency, testCommit, testCommit),
			consumer + "@" + testCommit:   migrationFetched(consumer, testCommit, testCommit, v1.Requirement{Module: dependency, Version: "v1.0.0"}),
			dependency + "@v2.0.0":        migrationFetched(dependency, "v2.0.0", testCommit),
			dependency + "@v1.0.0":        migrationFetched(dependency, "v1.0.0", strings.Repeat("a", 40)),
		},
		wantOldHashes: map[string]string{
			dependency + "@" + testCommit: testHash,
			consumer + "@" + testCommit:   testHash,
			dependency + "@v2.0.0":        "",
			dependency + "@v1.0.0":        "",
		},
	}
	lock, err := migrateLock(t.Context(), module, pins, true, repo)
	require.ErrorContains(t, err, "incompatible requirement example.com/acme/dep@v1.0.0")
	assert.Empty(t, lock.Modules)
}

func TestMigrateLockValidatesOnlyFinalSelectedMetadata(t *testing.T) {
	t.Parallel()
	const dependency = "example.com/acme/a"
	const consumer = "example.com/acme/b"
	const provisional = "example.com/acme/c"
	const selectedCommit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	module := v1.Module{Name: "example.com/acme/local", Requires: []v1.Requirement{{Module: dependency}, {Module: consumer}}}
	repo := &mockRepository{
		fetched: map[string]modules.Fetched{
			dependency + "@":               migrationFetched(dependency, testCommit, testCommit, v1.Requirement{Module: provisional, Version: "v0.0.0-20260101123456-" + testCommit}),
			consumer + "@":                 migrationFetched(consumer, testCommit, testCommit, v1.Requirement{Module: dependency, Version: "v1.0.0"}),
			provisional + "@" + testCommit: migrationFetched(provisional, testCommit, testCommit),
			dependency + "@v1.0.0":         migrationFetched(dependency, "v1.0.0", selectedCommit),
		},
		wantOldHashes: map[string]string{
			dependency + "@": "", consumer + "@": "",
			provisional + "@" + testCommit: "", dependency + "@v1.0.0": "",
		},
	}
	lock, err := migrateLock(t.Context(), module, nil, false, repo)
	require.NoError(t, err)
	require.Len(t, lock.Modules, 2)
	require.NoError(t, modules.ValidateRequirements(module.Requires, lock))
	for _, entry := range lock.Modules {
		assert.NotEqual(t, provisional, entry.Source)
		if entry.Source == dependency {
			assert.Equal(t, selectedCommit, entry.Commit)
			assert.Equal(t, "v1.0.0", entry.Version)
		}
	}
}

func migrationFetched(source, version, commit string, requirements ...v1.Requirement) modules.Fetched {
	return modules.Fetched{
		Module: v1.Module{Name: source, Roots: []string{"."}, Requires: requirements},
		Lock:   v1.LockedModule{Source: source, Version: version, Commit: commit, Hash: testNewHash},
	}
}

func TestLockMigration(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, version, resolved string }{
		{name: "sha", version: testCommit, resolved: testCommit},
		{name: "pseudo", version: "v0.0.0-20260101123456-" + testCommit, resolved: testCommit},
		{name: "tag", version: "v1.2.3", resolved: "v1.2.3"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFixture(t, root, "easyp.yaml", "deps: [example.com/acme/dep]\n")
			legacyLock := "example.com/acme/dep " + tt.version + " " + testHash + "\n"
			writeFixture(t, root, "easyp.lock", legacyLock)
			repo := &mockRepository{wantOldHash: testHash, fetched: map[string]modules.Fetched{
				"example.com/acme/dep@" + tt.resolved: {
					Module: v1.Module{Name: "example.com/acme/dep", Roots: []string{"."}},
					Lock:   v1.LockedModule{Source: "example.com/acme/dep", Version: tt.resolved, Commit: testCommit, Hash: testNewHash},
				},
			}}
			plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api", Repository: repo})
			require.NoError(t, err)
			assert.Empty(t, repo.calls)
			assert.NotEmpty(t, plan.Warnings())
			require.ErrorContains(t, plan.Apply(), "--resolve-lock")
			assert.Equal(t, "deps: [example.com/acme/dep]\n", string(mustRead(t, root, "easyp.yaml")))
			plan, err = Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api", ResolveLock: true, Repository: repo})
			require.NoError(t, err)
			var lock v1.Lock
			require.NoError(t, yaml.Unmarshal(outputContent(t, plan, "protobuf.lock"), &lock))
			require.Len(t, lock.Modules, 1)
			assert.Equal(t, testCommit, lock.Modules[0].Commit)
			assert.Equal(t, tt.resolved, lock.Modules[0].Version)
			assert.Equal(t, testNewHash, lock.Modules[0].Hash)
			require.NoError(t, plan.Apply())
			assert.Equal(t, legacyLock, string(mustRead(t, root, "easyp.lock")))
			calls := append([]string(nil), repo.calls...)
			second, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api", ResolveLock: true, Repository: repo})
			require.NoError(t, err)
			require.NoError(t, second.Apply())
			assert.Equal(t, calls, repo.calls, "repeat migration must not refresh")
		})
	}
}

func TestUnsafeLockMigration(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, old, errorText, requirement string
		mismatch                          bool
	}{
		{name: "missing_pin", old: "", errorText: "missing", requirement: ""},
		{name: "branch", old: "main", errorText: "full Git commit"},
		{name: "short_sha", old: "0123456", errorText: "full Git commit"},
		{name: "short_pseudo", old: "v0.0.0-20260101123456-0123456", errorText: "full Git commit"},
		{name: "hash_mismatch", old: testCommit, mismatch: true, errorText: "legacy hash mismatch"},
		{name: "transitive_missing", old: "v1.0.0", requirement: "other", errorText: "missing"},
		{name: "transitive_incompatible", old: "v1.0.0", requirement: "v2.0.0", errorText: "incompatible"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFixture(t, root, "easyp.yaml", "deps: [example.com/acme/dep]\n")
			lock := ""
			if tt.old != "" {
				lock = "example.com/acme/dep " + tt.old + " " + testHash + "\n"
			}
			writeFixture(t, root, "easyp.lock", lock)
			dep := v1.Module{Name: "example.com/acme/dep", Roots: []string{"."}}
			if tt.requirement == "other" {
				dep.Requires = []v1.Requirement{{Module: "example.com/acme/missing"}}
			} else if tt.requirement != "" {
				dep.Requires = []v1.Requirement{{Module: "example.com/acme/dep", Version: tt.requirement}}
			}
			repo := &mockRepository{wantOldHash: testHash, fetched: map[string]modules.Fetched{
				"example.com/acme/dep@" + tt.old: {Module: dep, Lock: v1.LockedModule{Source: "example.com/acme/dep", Version: tt.old, Commit: testCommit, Hash: testNewHash}},
			}}
			if tt.mismatch {
				repo.err = fmt.Errorf("legacy hash mismatch")
			}
			_, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api", ResolveLock: true, Repository: repo})
			require.ErrorContains(t, err, tt.errorText)
			_, err = os.Stat(filepath.Join(root, "easyp.yaml.v0.bak"))
			assert.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func TestLegacyManifest(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, manifest, wantError string }{
		{name: "blocks_and_inline", manifest: "direct (\nexample.com/acme/a@v1.2.3\n)\nindirect (\nexample.com/acme/b@" + testCommit + "\n)\nexample.com/acme/c@v2.0.0\n"},
		{name: "local_replacement", manifest: "direct (\nexample.com/acme/a@v1.2.3\n)\nreplace example.com/acme/a@v1.2.3 => ../a\n"},
		{name: "ambiguous_replace", manifest: "direct example.com/acme/a@v1.2.3\nreplace example.com/acme/a@v1.2.3 => ../a\nreplace example.com/acme/a@v2.0.0 => ../b\n", wantError: "ambiguous"},
		{name: "wrong_replace", manifest: "direct example.com/acme/a@v1.2.3\nreplace example.com/acme/a@v2.0.0 => ../a\n", wantError: "version-specific"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFixture(t, root, "easyp.yaml", "lint: {}\n")
			writeFixture(t, root, "protobuf.mod", tt.manifest)
			plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api"})
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			candidate := string(outputContent(t, plan, "protobuf.mod"))
			assert.Contains(t, candidate, "require example.com/acme/a v1.2.3")
			if strings.Contains(tt.manifest, "indirect") {
				assert.Contains(t, candidate, " // indirect")
			}
			if strings.Contains(tt.manifest, "replace") {
				assert.Contains(t, candidate, "replace example.com/acme/a => ../a")
			}
			assert.Equal(t, tt.manifest, string(mustRead(t, root, "protobuf.mod")))
		})
	}
}
