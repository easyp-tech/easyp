package modules_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

func TestTidyUnusedDuplicatesKeepIntrinsicRootInference(t *testing.T) {
	t.Parallel()
	fixture := newTidyCollisionFixture(t, "pinned git", "", map[string]string{"api/unused.proto": "syntax = \"proto3\"; package first;"}, map[string]string{"proto/unused.proto": "syntax = \"proto3\"; package second;"})
	provider, commit := importRootsRepository(t, map[string]string{
		"api/request.proto": "syntax = \"proto3\"; import \"extra.proto\";",
		"api/extra.proto":   "syntax = \"proto3\"; package extra; message Extra {}",
	}, nil)
	manifest := string(importRootsRead(t, fixture.consumer, v1.ModuleFile)) + "require " + provider + " " + commit + "\n"
	importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
	importRootsWrite(t, fixture.consumer, "api/extra_client.proto", "syntax = \"proto3\"; import \"extra.proto\"; message ExtraClient { extra.Extra value = 1; }")

	err := modules.Tidy(t.Context(), fixture.consumer, fixture.cache)

	require.NoError(t, err)
	lock, err := modules.ReadLock(filepath.Join(fixture.consumer, v1.LockFile))
	require.NoError(t, err)
	require.Len(t, lock.Modules, 3)
	found := false
	for _, entry := range lock.Modules {
		if entry.Source == provider {
			found = true
			assert.Equal(t, []string{"api"}, entry.Roots)
			assert.Equal(t, commit, entry.Commit)
		}
	}
	assert.True(t, found)
}

func TestTidyDuplicateCandidatesRespectBufExclusions(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"pinned git", "local replace"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			fixture := newTidyCollisionFixture(t, mode, "", map[string]string{"api/common/v1/types.proto": strings.Replace(tidyCommonProto, "string value", "bool value", 1)}, nil)
			require.NoError(t, os.Remove(filepath.Join(fixture.producer, v1.ModuleFile)))
			importRootsWrite(t, fixture.producer, "buf.yaml", "version: v2\nmodules:\n  - path: api\n    excludes: [api/common]\n")
			importRootsWrite(t, fixture.producer, "easyp.yaml", "deps: ["+fixture.common+"@"+fixture.commits[fixture.common]+"]\ngenerate:\n  inputs: [{directory: {root: api}}]\n")
			commit := tidyCollisionCommit(t, fixture.producer)
			manifest := strings.Replace(string(importRootsRead(t, fixture.consumer, v1.ModuleFile)), fixture.commits[fixture.producer], commit, 1)
			importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)

			err := modules.Tidy(t.Context(), fixture.consumer, fixture.cache)

			require.NoError(t, err)
			assert.Equal(t, tidyCommonProto, string(importRootsRead(t, fixture.common, "proto/common/v1/types.proto")))
		})
	}
}

func TestTidyDuplicateCandidatesRecheckAliasIdentity(t *testing.T) {
	t.Parallel()
	fixture := newTidyCollisionFixture(t, "local replace", "", map[string]string{"api/materialized/one.proto": tidyCommonProto, "api/materialized/two.proto": tidyCommonProto}, nil)
	alias := filepath.Join(fixture.producer, "api/common/v1/types.proto")
	require.NoError(t, os.MkdirAll(filepath.Dir(alias), 0o755))
	require.NoError(t, os.Symlink("../../materialized/one.proto", alias))
	manifest := string(importRootsRead(t, fixture.consumer, v1.ModuleFile))
	manifest = strings.Replace(manifest, "replace "+fixture.common+" => "+fixture.common+"\n", "", 1)
	importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
	importRootsWrite(t, fixture.consumer, v1.LockFile, "version: 1\nmodules: []\n")
	require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))
	before := importRootsRead(t, fixture.consumer, v1.LockFile)
	spy := &tidyCollisionMutation{Cache: fixture.cache, at: 1, mutate: func(v1.Lock) {
		require.NoError(t, os.Remove(alias))
		require.NoError(t, os.Symlink("../../materialized/two.proto", alias))
	}}

	err := modules.Tidy(t.Context(), fixture.consumer, spy)

	require.ErrorIs(t, err, sourceview.ErrChanged)
	assert.Equal(t, manifest, string(importRootsRead(t, fixture.consumer, v1.ModuleFile)))
	assert.Equal(t, before, importRootsRead(t, fixture.consumer, v1.LockFile))
}

func TestTidyConflictingConsumerTargets(t *testing.T) {
	t.Parallel()
	for _, different := range []bool{false, true} {
		t.Run(map[bool]string{false: "equal", true: "different"}[different], func(t *testing.T) {
			t.Parallel()
			fixture := newTidyCollisionFixture(t, "pinned git", "", nil, nil)
			manifest := strings.Replace(string(importRootsRead(t, fixture.consumer, v1.ModuleFile)), "roots api\n", "roots api other\n", 1)
			importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
			copy := string(importRootsRead(t, fixture.consumer, "api/consumer/request.proto"))
			if different {
				copy += "// distinct consumer contract\n"
			}
			importRootsWrite(t, fixture.consumer, "other/consumer/request.proto", copy)

			err := modules.Tidy(t.Context(), fixture.consumer, fixture.cache)

			if different {
				require.ErrorContains(t, err, "duplicate import path")
				assert.ErrorContains(t, err, "consumer/request.proto")
				assert.Equal(t, manifest, string(importRootsRead(t, fixture.consumer, v1.ModuleFile)))
				assert.NoFileExists(t, filepath.Join(fixture.consumer, v1.LockFile))
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTidyIdenticalCopiesKeepPinsOnColdCache(t *testing.T) {
	t.Parallel()
	fixture := newTidyCollisionFixture(t, "pinned git", "", map[string]string{"api/common/v1/types.proto": tidyCommonProto}, nil)
	require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))
	manifest, lock := importRootsRead(t, fixture.consumer, v1.ModuleFile), importRootsRead(t, fixture.consumer, v1.LockFile)

	err := modules.Tidy(t.Context(), fixture.consumer, gitmodules.New(t.TempDir()))

	require.NoError(t, err)
	assert.Equal(t, manifest, importRootsRead(t, fixture.consumer, v1.ModuleFile))
	assert.Equal(t, lock, importRootsRead(t, fixture.consumer, v1.LockFile))
}

func TestTidyEqualCopiesPromoteOnlyCanonicalProvider(t *testing.T) {
	t.Parallel()
	fixture := newTidyCollisionFixture(t, "pinned git", "", map[string]string{"api/common/v1/types.proto": tidyCommonProto}, nil)
	manifest := string(importRootsRead(t, fixture.consumer, v1.ModuleFile))
	for _, commit := range fixture.commits {
		manifest = strings.Replace(manifest, commit+"\n", commit+" // indirect\n", 1)
	}
	importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
	importRootsWrite(t, fixture.consumer, "api/consumer/request.proto", "syntax = \"proto3\"; import \"common/v1/types.proto\"; message Request { common.v1.Common value = 1; }")

	err := modules.Tidy(t.Context(), fixture.consumer, fixture.cache)

	require.NoError(t, err)
	updated := string(importRootsRead(t, fixture.consumer, v1.ModuleFile))
	assert.Contains(t, updated, "require "+fixture.producer+" "+fixture.commits[fixture.producer]+" // indirect\n")
	assert.Contains(t, updated, "require "+fixture.common+" "+fixture.commits[fixture.common]+"\n")
}

func TestTidyRenamePreservesCanonicalSharedProvider(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		consumerCopy  bool
		producerFirst bool
		wantRewrites  int
	}{
		{name: "unchanged dependency owns the import", wantRewrites: 1},
		{name: "consumer owns the import", consumerCopy: true, wantRewrites: 1},
		{name: "updated dependency owned the import", producerFirst: true, wantRewrites: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixture := newTidyCollisionFixture(t, "pinned git", "", map[string]string{"api/common/v1/types.proto": tidyCommonProto}, nil)
			if tt.consumerCopy {
				importRootsWrite(t, fixture.consumer, "api/common/v1/types.proto", tidyCommonProto)
			}
			require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))
			if tt.producerFirst {
				// A valid lock's module order is the previous source precedence.
				// Put the producer's copy first to exercise the opposite owner.
				lock, err := modules.ReadLock(filepath.Join(fixture.consumer, v1.LockFile))
				require.NoError(t, err)
				slices.Reverse(lock.Modules)
				raw, err := yaml.Marshal(lock)
				require.NoError(t, err)
				importRootsWrite(t, fixture.consumer, v1.LockFile, string(raw))
				service := string(importRootsRead(t, fixture.producer, "api/service/v1/service.proto"))
				importRootsWrite(t, fixture.producer, "api/service/v1/service.proto", strings.Replace(service, "common/v1/types.proto", "v1/types.proto", 1))
			}
			consumer := string(importRootsRead(t, fixture.consumer, "api/consumer/request.proto"))
			manifest := string(importRootsRead(t, fixture.producer, v1.ModuleFile))
			manifest = strings.Replace(manifest, "roots api\n", "roots api/service api/common\n", 1)
			importRootsWrite(t, fixture.producer, v1.ModuleFile, manifest)
			commit := tidyCollisionCommit(t, fixture.producer)
			manifest = strings.Replace(string(importRootsRead(t, fixture.consumer, v1.ModuleFile)), fixture.commits[fixture.producer], commit, 1)
			importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)

			report, err := modules.TidyWithReport(t.Context(), fixture.consumer, fixture.cache)

			require.NoError(t, err)
			require.Len(t, report.Imports, tt.wantRewrites)
			want := strings.Replace(consumer, "service/v1/service.proto", "v1/service.proto", 1)
			if tt.producerFirst {
				want = strings.Replace(want, "common/v1/types.proto", "v1/types.proto", 1)
			}
			assert.Equal(t, want, string(importRootsRead(t, fixture.consumer, "api/consumer/request.proto")))
		})
	}
}

func TestTidyOverlayCapturesMetadataBeforeRemoteAcquisition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		initial map[string]string
		change  map[string]string
	}{
		{name: "native roots", change: map[string]string{v1.ModuleFile: "roots other"}},
		{name: "Buf workspace member", initial: map[string]string{
			"buf.work.yaml": "version: v1\ndirectories: [api]\n",
			"api/buf.yaml":  "version: v1\n",
		}, change: map[string]string{"api/buf.yaml": "version: v1\nbuild:\n  excludes: [common]\n"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixture := newTidyCollisionFixture(t, "local replace", "", map[string]string{"api/common/v1/types.proto": tidyCommonProto}, nil)
			manifest := string(importRootsRead(t, fixture.consumer, v1.ModuleFile))
			manifest = strings.Replace(manifest, "replace "+fixture.common+" => "+fixture.common+"\n", "", 1)
			importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
			if len(tt.initial) > 0 {
				require.NoError(t, os.Remove(filepath.Join(fixture.producer, v1.ModuleFile)))
				importRootsWrite(t, fixture.producer, "easyp.yaml", "deps: ["+fixture.common+"@"+fixture.commits[fixture.common]+"]\n")
				for name, raw := range tt.initial {
					importRootsWrite(t, fixture.producer, name, raw)
				}
			} else {
				importRootsWrite(t, fixture.producer, "other/service/v1/service.proto", string(importRootsRead(t, fixture.producer, "api/service/v1/service.proto")))
			}
			consumer := importRootsRead(t, fixture.consumer, "api/consumer/request.proto")
			spy := &tidyOverlayAcquisitionMutation{Cache: fixture.cache, mutate: func() {
				for name, raw := range tt.change {
					if name == v1.ModuleFile {
						raw = strings.Replace(string(importRootsRead(t, fixture.producer, name)), "roots api", raw, 1)
					}
					importRootsWrite(t, fixture.producer, name, raw)
				}
			}}

			err := modules.Tidy(t.Context(), fixture.consumer, spy)

			require.True(t, spy.changed, "mutation must happen during remote acquisition")
			require.ErrorIs(t, err, sourceview.ErrChanged)
			assert.Equal(t, manifest, string(importRootsRead(t, fixture.consumer, v1.ModuleFile)))
			assert.Equal(t, consumer, importRootsRead(t, fixture.consumer, "api/consumer/request.proto"))
			assert.NoFileExists(t, filepath.Join(fixture.consumer, v1.LockFile))
		})
	}
}

type tidyOverlayAcquisitionMutation struct {
	*gitmodules.Cache
	mutate  func()
	changed bool
}

func (repository *tidyOverlayAcquisitionMutation) Install(ctx context.Context, lock v1.Lock) error {
	if !repository.changed && len(lock.Modules) > 0 {
		repository.changed = true
		repository.mutate()
	}
	return repository.Cache.Install(ctx, lock)
}

func TestTidyOverlayWithRelativeCacheDirectory(t *testing.T) {
	// t.Chdir excludes parallel tests while the relative cache is in use.
	t.Chdir(t.TempDir())
	fixture := newTidyCollisionFixture(t, "local replace", "", map[string]string{"api/common/v1/types.proto": tidyCommonProto}, nil)
	manifest := string(importRootsRead(t, fixture.consumer, v1.ModuleFile))
	manifest = strings.Replace(manifest, "replace "+fixture.common+" => "+fixture.common+"\n", "", 1)
	importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)

	err := modules.Tidy(t.Context(), fixture.consumer, gitmodules.New("cache"))

	require.NoError(t, err)
	assert.Equal(t, manifest, string(importRootsRead(t, fixture.consumer, v1.ModuleFile)))
	assert.NoFileExists(t, filepath.Join(fixture.consumer, v1.LockFile))
}

func TestTidyOverlayRechecksIntrinsicRootEvidence(t *testing.T) {
	t.Parallel()
	for _, mutation := range []string{"change unused inference source", "add new intrinsic root", "add nested producer metadata"} {
		t.Run(mutation, func(t *testing.T) {
			t.Parallel()
			fixture := newTidyCollisionFixture(t, "local replace", "", map[string]string{"api/common/v1/types.proto": tidyCommonProto}, nil)
			manifest := string(importRootsRead(t, fixture.consumer, v1.ModuleFile))
			manifest = strings.Replace(manifest, "replace "+fixture.common+" => "+fixture.common+"\n", "", 1)
			importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
			require.NoError(t, os.Remove(filepath.Join(fixture.producer, v1.ModuleFile)))
			importRootsWrite(t, fixture.producer, "api/request.proto", "syntax = \"proto3\"; import \"svc.proto\";")
			importRootsWrite(t, fixture.producer, "api/svc.proto", "syntax = \"proto3\"; package service; message Service {}")
			const consumer = "syntax = \"proto3\"; import \"svc.proto\"; import \"common/v1/types.proto\"; message Request { service.Service service = 1; common.v1.Common common = 2; }"
			importRootsWrite(t, fixture.consumer, "api/consumer/request.proto", consumer)
			spy := &tidyCollisionMutation{Cache: fixture.cache, at: 1, mutate: func(v1.Lock) {
				if mutation == "change unused inference source" {
					importRootsWrite(t, fixture.producer, "api/request.proto", "syntax = \"proto3\"; import \"api/svc.proto\";")
					return
				}
				if mutation == "add nested producer metadata" {
					importRootsWrite(t, fixture.producer, "newmodule/protobuf.mod", "module "+fixture.producer+"\nroots .\n")
					return
				}
				importRootsWrite(t, fixture.producer, "schema/request.proto", "syntax = \"proto3\"; import \"svc.proto\";")
				importRootsWrite(t, fixture.producer, "schema/svc.proto", "syntax = \"proto3\"; package other; message Other {}")
			}}

			err := modules.Tidy(t.Context(), fixture.consumer, spy)

			require.Equal(t, 1, spy.checks, "mutation must happen after local root inference")
			require.ErrorIs(t, err, sourceview.ErrChanged)
			assert.Equal(t, manifest, string(importRootsRead(t, fixture.consumer, v1.ModuleFile)))
			assert.Equal(t, consumer, string(importRootsRead(t, fixture.consumer, "api/consumer/request.proto")))
			assert.NoFileExists(t, filepath.Join(fixture.consumer, v1.LockFile))
		})
	}
}

func TestTidyOverlayAcceptsImportRootDirectoryAlias(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"native", "Buf"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			fixture := newTidyCollisionFixture(t, "local replace", "", nil, nil)
			require.NoError(t, os.Symlink("api", filepath.Join(fixture.producer, "proto")))
			if mode == "native" {
				manifest := string(importRootsRead(t, fixture.producer, v1.ModuleFile))
				importRootsWrite(t, fixture.producer, v1.ModuleFile, strings.Replace(manifest, "roots api", "roots proto", 1))
			} else {
				require.NoError(t, os.Remove(filepath.Join(fixture.producer, v1.ModuleFile)))
				importRootsWrite(t, fixture.producer, "buf.yaml", "version: v2\nmodules: [{path: proto}]\n")
				importRootsWrite(t, fixture.producer, "easyp.yaml", "deps: ["+fixture.common+"@"+fixture.commits[fixture.common]+"]\n")
			}

			err := modules.Tidy(t.Context(), fixture.consumer, fixture.cache)

			require.NoError(t, err)
			assert.NoFileExists(t, filepath.Join(fixture.consumer, v1.LockFile))
		})
	}
}

func TestTidyOverlayRechecksAbsentMetadataThroughDirectoryAlias(t *testing.T) {
	t.Parallel()
	fixture := newTidyCollisionFixture(t, "local replace", "", nil, nil)
	manifest := strings.Replace(string(importRootsRead(t, fixture.consumer, v1.ModuleFile)), "replace "+fixture.common+" => "+fixture.common+"\n", "", 1)
	importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
	producer := string(importRootsRead(t, fixture.producer, v1.ModuleFile))
	importRootsWrite(t, fixture.producer, v1.ModuleFile, strings.Replace(producer, "roots api", "roots proto", 1))
	require.NoError(t, os.Symlink("api", filepath.Join(fixture.producer, "proto")))
	importRootsWrite(t, fixture.producer, "other/service/v1/service.proto", string(importRootsRead(t, fixture.producer, "api/service/v1/service.proto")))
	spy := &tidyCollisionMutation{Cache: fixture.cache, at: 1, mutate: func(v1.Lock) {
		require.NoError(t, os.Remove(filepath.Join(fixture.producer, "proto")))
		require.NoError(t, os.Symlink("other", filepath.Join(fixture.producer, "proto")))
	}}

	err := modules.Tidy(t.Context(), fixture.consumer, spy)

	require.Equal(t, 1, spy.checks, "mutation must happen after absent metadata capture")
	require.ErrorIs(t, err, sourceview.ErrChanged)
	assert.Equal(t, manifest, string(importRootsRead(t, fixture.consumer, v1.ModuleFile)))
	assert.NoFileExists(t, filepath.Join(fixture.consumer, v1.LockFile))
}

func TestTidyOverlayRejectsDanglingMetadataFileAlias(t *testing.T) {
	t.Parallel()
	fixture := newTidyCollisionFixture(t, "local replace", "", nil, nil)
	require.NoError(t, os.Symlink("missing.yaml", filepath.Join(fixture.producer, "api/easyp.yaml")))

	err := modules.Tidy(t.Context(), fixture.consumer, fixture.cache)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "api/easyp.yaml")
	assert.NoFileExists(t, filepath.Join(fixture.consumer, v1.LockFile))
}
