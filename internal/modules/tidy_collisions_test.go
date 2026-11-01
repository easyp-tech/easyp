package modules_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

const tidyCommonProto = "syntax = \"proto3\"; package common.v1; message Common { string value = 1; }\n"

func TestTidyDuplicateImportClosure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		copies       map[string]string
		common       map[string]string
		consumerCopy string
		roots        string
		wantConflict string
	}{
		{name: "one direct and transitive provider"},
		{name: "unused identical duplicates", copies: map[string]string{"api/tests/unused.proto": "syntax = \"proto3\"; package unused;"}, common: map[string]string{"proto/tests/unused.proto": "syntax = \"proto3\"; package unused;"}},
		{name: "unused differing duplicates", copies: map[string]string{"api/tests/unused.proto": "syntax = \"proto3\"; package first;"}, common: map[string]string{"proto/tests/unused.proto": "syntax = \"proto3\"; package second;"}},
		{name: "unused malformed duplicate", copies: map[string]string{"api/tests/unused.proto": "this source is never imported"}, common: map[string]string{"proto/tests/unused.proto": "syntax = \"proto3\"; package unused;"}},
		{name: "reachable identical copies", copies: map[string]string{"api/common/v1/types.proto": tidyCommonProto}},
		{name: "reachable differing copies", copies: map[string]string{"api/common/v1/types.proto": strings.Replace(tidyCommonProto, "string value", "int64 value", 1)}, wantConflict: "common/v1/types.proto"},
		{name: "reachable comment difference", copies: map[string]string{"api/common/v1/types.proto": tidyCommonProto + "// another contract source\n"}, wantConflict: "common/v1/types.proto"},
		{name: "identical copies in another root", roots: "api extra", copies: map[string]string{"extra/common/v1/types.proto": tidyCommonProto}},
		{name: "conflicting copies in another root", roots: "api extra", copies: map[string]string{"extra/common/v1/types.proto": strings.Replace(tidyCommonProto, "string value", "bytes value", 1)}, wantConflict: "common/v1/types.proto"},
		{name: "identical consumer copy", consumerCopy: tidyCommonProto},
		{name: "conflicting consumer copy", consumerCopy: strings.Replace(tidyCommonProto, "string value", "bool value", 1), wantConflict: "common/v1/types.proto"},
	}
	for _, mode := range []string{"pinned git", "local replace"} {
		for _, tt := range tests {
			t.Run(mode+"/"+tt.name, func(t *testing.T) {
				t.Parallel()
				fixture := newTidyCollisionFixture(t, mode, tt.roots, tt.copies, tt.common)
				if tt.consumerCopy != "" {
					importRootsWrite(t, fixture.consumer, "api/common/v1/types.proto", tt.consumerCopy)
				}
				beforeManifest := importRootsRead(t, fixture.consumer, v1.ModuleFile)
				beforeSource := importRootsRead(t, fixture.consumer, "api/consumer/request.proto")
				beforeProducer := importRootsRead(t, fixture.producer, "api/service/v1/service.proto")
				beforeCommon := importRootsRead(t, fixture.common, "proto/common/v1/types.proto")
				// Even invalid/old lock bytes must remain untouched in replacement mode.
				const localLock = "version: 1\nmodules: []\n# local developer lock\n"
				if mode == "local replace" {
					importRootsWrite(t, fixture.consumer, v1.LockFile, localLock)
				}

				err := modules.Tidy(t.Context(), fixture.consumer, fixture.cache)

				assert.Equal(t, beforeSource, importRootsRead(t, fixture.consumer, "api/consumer/request.proto"))
				assert.Equal(t, beforeProducer, importRootsRead(t, fixture.producer, "api/service/v1/service.proto"))
				assert.Equal(t, beforeCommon, importRootsRead(t, fixture.common, "proto/common/v1/types.proto"))
				if tt.wantConflict != "" {
					require.ErrorContains(t, err, "duplicate import path")
					assert.ErrorContains(t, err, tt.wantConflict)
					assert.Equal(t, beforeManifest, importRootsRead(t, fixture.consumer, v1.ModuleFile))
					if mode == "pinned git" {
						assert.NoFileExists(t, filepath.Join(fixture.consumer, v1.LockFile))
					}
				} else {
					require.NoError(t, err)
					if mode == "pinned git" {
						lock, err := modules.ReadLock(filepath.Join(fixture.consumer, v1.LockFile))
						require.NoError(t, err)
						require.Len(t, lock.Modules, 2)
						for _, entry := range lock.Modules {
							assert.Equal(t, fixture.commits[entry.Source], entry.Commit)
							assert.Equal(t, entry.Commit, entry.Version)
						}
					}
					manifest, lock := importRootsRead(t, fixture.consumer, v1.ModuleFile), importRootsRead(t, fixture.consumer, v1.LockFile)
					require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))
					assert.Equal(t, manifest, importRootsRead(t, fixture.consumer, v1.ModuleFile))
					assert.Equal(t, lock, importRootsRead(t, fixture.consumer, v1.LockFile))
				}
				if mode == "local replace" {
					assert.Equal(t, beforeManifest, importRootsRead(t, fixture.consumer, v1.ModuleFile))
					assert.Equal(t, localLock, string(importRootsRead(t, fixture.consumer, v1.LockFile)))
				}
			})
		}
	}
}

type tidyCollisionFixture struct {
	consumer string
	producer string
	common   string
	commits  map[string]string
	cache    *gitmodules.Cache
}

func newTidyCollisionFixture(t *testing.T, mode, roots string, copies, commonFiles map[string]string) tidyCollisionFixture {
	t.Helper()
	common := map[string]string{"proto/common/v1/types.proto": tidyCommonProto}
	for name, content := range commonFiles {
		common[name] = content
	}
	commonRepo, _ := importRootsRepository(t, common, nil)
	importRootsWrite(t, commonRepo, v1.ModuleFile, "module "+commonRepo+"\nroots proto\n")
	commonCommit := tidyCollisionCommit(t, commonRepo)
	service := "syntax = \"proto3\"; package service.v1; import public \"common/v1/types.proto\"; message Service { common.v1.Common common = 1; }\n"
	files := map[string]string{"api/service/v1/service.proto": service}
	for name, content := range copies {
		files[name] = content
	}
	producer, _ := importRootsRepository(t, files, nil)
	if roots == "" {
		roots = "api"
	}
	importRootsWrite(t, producer, v1.ModuleFile, "module "+producer+"\nroots "+roots+"\nrequire "+commonRepo+" "+commonCommit+"\n")
	producerCommit := tidyCollisionCommit(t, producer)
	root := t.TempDir()
	manifest := "module example.test/consumer\nroots api\nrequire " + producer + " " + producerCommit + "\nrequire " + commonRepo + " " + commonCommit + "\n"
	if mode == "local replace" {
		manifest += "replace " + producer + " => " + producer + "\nreplace " + commonRepo + " => " + commonRepo + "\n"
	}
	importRootsWrite(t, root, v1.ModuleFile, manifest)
	importRootsWrite(t, root, "api/consumer/request.proto", "syntax = \"proto3\"; package consumer; import \"service/v1/service.proto\"; import \"common/v1/types.proto\"; message Request { service.v1.Service service = 1; common.v1.Common common = 2; }\n")
	importRootsWrite(t, root, v1.GenerateFile, "version: v1\nplugins: [{name: python, out: gen/python}]\n")
	return tidyCollisionFixture{consumer: root, producer: producer, common: commonRepo, commits: map[string]string{producer: producerCommit, commonRepo: commonCommit}, cache: gitmodules.New(t.TempDir())}
}

func tidyCollisionCommit(t *testing.T, directory string) string {
	t.Helper()
	importRootsGit(t, directory, "add", ".")
	importRootsGit(t, directory, "commit", "--quiet", "-m", "dependency fixture metadata")
	return importRootsGit(t, directory, "rev-parse", "HEAD")
}

func TestTidyComparedProtoClosure(t *testing.T) {
	t.Parallel()
	for _, conflict := range []bool{false, true} {
		t.Run(map[bool]string{false: "identical transitive copies", true: "different transitive copies"}[conflict], func(t *testing.T) {
			t.Parallel()
			parent := "syntax = \"proto3\"; package common.v1; import weak \"common/v1/detail.proto\"; message Common { Detail value = 1; }\n"
			detail := "syntax = \"proto3\"; package common.v1; message Detail { string value = 1; }\n"
			copiedDetail := detail
			if conflict {
				copiedDetail = strings.Replace(detail, "string value", "bytes value", 1)
			}
			fixture := newTidyCollisionFixture(t, "pinned git", "", map[string]string{"api/common/v1/types.proto": parent, "api/common/v1/detail.proto": copiedDetail}, map[string]string{"proto/common/v1/types.proto": parent, "proto/common/v1/detail.proto": detail})
			before := importRootsRead(t, fixture.consumer, v1.ModuleFile)

			err := modules.Tidy(t.Context(), fixture.consumer, fixture.cache)

			if conflict {
				require.ErrorContains(t, err, "duplicate import path")
				assert.ErrorContains(t, err, "common/v1/detail.proto")
				assert.Equal(t, before, importRootsRead(t, fixture.consumer, v1.ModuleFile))
				assert.NoFileExists(t, filepath.Join(fixture.consumer, v1.LockFile))
				return
			}
			require.NoError(t, err)
			lock, err := modules.ReadLock(filepath.Join(fixture.consumer, v1.LockFile))
			require.NoError(t, err)
			require.Len(t, lock.Modules, 2)
		})
	}
}

func TestTidyDuplicateConflictPreservesExistingLock(t *testing.T) {
	t.Parallel()
	fixture := newTidyCollisionFixture(t, "pinned git", "", nil, nil)
	require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))
	importRootsWrite(t, fixture.producer, "api/common/v1/types.proto", strings.Replace(tidyCommonProto, "string value", "int32 value", 1))
	commit := tidyCollisionCommit(t, fixture.producer)
	manifest := string(importRootsRead(t, fixture.consumer, v1.ModuleFile))
	manifest = strings.Replace(manifest, fixture.commits[fixture.producer], commit, 1)
	importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
	lock, source := importRootsRead(t, fixture.consumer, v1.LockFile), importRootsRead(t, fixture.consumer, "api/consumer/request.proto")

	err := modules.Tidy(t.Context(), fixture.consumer, fixture.cache)

	require.ErrorContains(t, err, "duplicate import path")
	assert.Equal(t, manifest, string(importRootsRead(t, fixture.consumer, v1.ModuleFile)))
	assert.Equal(t, lock, importRootsRead(t, fixture.consumer, v1.LockFile))
	assert.Equal(t, source, importRootsRead(t, fixture.consumer, "api/consumer/request.proto"))
	_, err = os.Stat(filepath.Join(fixture.consumer, v1.LockFile))
	require.NoError(t, err)
}

func TestTidyRenamesImportWithUnusedDuplicateInventory(t *testing.T) {
	t.Parallel()
	for _, inventory := range []string{"old", "new"} {
		t.Run(inventory, func(t *testing.T) {
			t.Parallel()
			roots, nextRoots := "api extra", "api/service extra"
			if inventory == "new" {
				roots, nextRoots = nextRoots, roots
			}
			fixture := newTidyCollisionFixture(t, "pinned git", roots, map[string]string{
				"api/tests/unused.proto":   "syntax = \"proto3\"; package first;",
				"extra/tests/unused.proto": "syntax = \"proto3\"; package second;",
			}, nil)
			consumer := string(importRootsRead(t, fixture.consumer, "api/consumer/request.proto"))
			beforeName, afterName := "service/v1/service.proto", "v1/service.proto"
			if inventory == "new" {
				beforeName, afterName = afterName, beforeName
			}
			consumer = strings.Replace(consumer, "service/v1/service.proto", beforeName, 1)
			importRootsWrite(t, fixture.consumer, "api/consumer/request.proto", consumer)
			require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))
			manifest := string(importRootsRead(t, fixture.producer, v1.ModuleFile))
			manifest = strings.Replace(manifest, "roots "+roots+"\n", "roots "+nextRoots+"\n", 1)
			importRootsWrite(t, fixture.producer, v1.ModuleFile, manifest)
			commit := tidyCollisionCommit(t, fixture.producer)
			manifest = strings.Replace(string(importRootsRead(t, fixture.consumer, v1.ModuleFile)), fixture.commits[fixture.producer], commit, 1)
			importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)

			report, err := modules.TidyWithReport(t.Context(), fixture.consumer, fixture.cache)

			require.NoError(t, err)
			require.Len(t, report.Imports, 1)
			assert.Equal(t, beforeName, report.Imports[0].From)
			assert.Equal(t, afterName, report.Imports[0].To)
			assert.Equal(t, strings.Replace(consumer, beforeName, afterName, 1), string(importRootsRead(t, fixture.consumer, "api/consumer/request.proto")))
			assert.Equal(t, manifest, string(importRootsRead(t, fixture.consumer, v1.ModuleFile)))
		})
	}
}

type tidyCollisionMutation struct {
	*gitmodules.Cache
	checks int
	at     int
	mutate func(v1.Lock)
}

func (cache *tidyCollisionMutation) VerifyCached(ctx context.Context, lock v1.Lock) error {
	cache.checks++
	err := cache.Cache.VerifyCached(ctx, lock)
	if err == nil && cache.checks == cache.at {
		cache.mutate(lock)
	}
	return err
}

func TestTidyDuplicateComparisonRechecksInputs(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"pinned git", "mixed overlay", "mixed overlay remote copy"} {
		for _, mutation := range []string{"changed copy", "added copy", "added differing copy"} {
			t.Run(mode+"/"+mutation, func(t *testing.T) {
				t.Parallel()
				copies := map[string]string{}
				if mutation == "changed copy" {
					copies["api/common/v1/types.proto"] = tidyCommonProto
				}
				fixtureMode := mode
				if mode != "pinned git" {
					fixtureMode = "local replace"
				}
				fixture := newTidyCollisionFixture(t, fixtureMode, "", copies, nil)
				if mode != "pinned git" {
					manifest := string(importRootsRead(t, fixture.consumer, v1.ModuleFile))
					remote := fixture.common
					if mode == "mixed overlay remote copy" {
						remote = fixture.producer
					}
					manifest = strings.Replace(manifest, "replace "+remote+" => "+remote+"\n", "", 1)
					importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
					importRootsWrite(t, fixture.consumer, v1.LockFile, "version: 1\nmodules: []\n")
				}
				require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))
				manifest, lock := importRootsRead(t, fixture.consumer, v1.ModuleFile), importRootsRead(t, fixture.consumer, v1.LockFile)
				source := importRootsRead(t, fixture.consumer, "api/consumer/request.proto")
				at := 2
				if mode != "pinned git" {
					at = 1
				}
				spy := &tidyCollisionMutation{Cache: fixture.cache, at: at, mutate: func(current v1.Lock) {
					directory := fixture.producer
					if mode != "mixed overlay" {
						for _, entry := range current.Modules {
							if entry.Source == fixture.producer {
								var err error
								directory, _, err = fixture.cache.Cached(entry)
								require.NoError(t, err)
							}
						}
					}
					data := tidyCommonProto
					if mutation != "added copy" {
						data = strings.Replace(data, "string value", "bool value", 1)
					}
					importRootsWrite(t, directory, "api/common/v1/types.proto", data)
				}}

				err := modules.Tidy(t.Context(), fixture.consumer, spy)

				require.ErrorIs(t, err, sourceview.ErrChanged)
				assert.GreaterOrEqual(t, spy.checks, at)
				assert.Equal(t, manifest, importRootsRead(t, fixture.consumer, v1.ModuleFile))
				assert.Equal(t, lock, importRootsRead(t, fixture.consumer, v1.LockFile))
				assert.Equal(t, source, importRootsRead(t, fixture.consumer, "api/consumer/request.proto"))
			})
		}
	}
}
