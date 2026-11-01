package modules_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

const tidyHistoricalCurrentProto = "syntax = \"proto3\"; package model; message New { string value = 1; }\n"
const tidyHistoricalConsumerProto = "syntax = \"proto3\"; package consumer; import \"new/v1/model.proto\"; message Request { model.New value = 1; }\n"

func TestTidyHistoricalCollisionsRetainUnambiguousCurrentImport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		conflictingRoots bool
		retainProducer   bool
	}{
		{name: "remove conflicting provider"},
		{name: "remove provider with conflicting roots", conflictingRoots: true},
		{name: "current copies have identical bytes", retainProducer: true},
	}
	for _, tt := range tests {
		for _, reverse := range []bool{false, true} {
			name := tt.name + "/original lock order"
			if reverse {
				name = tt.name + "/reversed lock order"
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				fixture := newTidyHistoricalCollisionFixture(t, tt.conflictingRoots, reverse)
				manifest := "module example.test/consumer\nroots api\nrequire " + fixture.common + " " + fixture.commits[fixture.common] + "\n"
				if tt.retainProducer {
					importRootsWrite(t, fixture.producer, "api/new/v1/model.proto", tidyHistoricalCurrentProto)
					commit := tidyCollisionCommit(t, fixture.producer)
					manifest += "require " + fixture.producer + " " + commit + "\n"
				}
				importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
				importRootsWrite(t, fixture.consumer, "api/consumer/request.proto", tidyHistoricalConsumerProto)

				// The same current project is valid without historical ambiguity.
				fresh := t.TempDir()
				importRootsWrite(t, fresh, v1.ModuleFile, manifest)
				importRootsWrite(t, fresh, "api/consumer/request.proto", tidyHistoricalConsumerProto)
				require.NoError(t, modules.Tidy(t.Context(), fresh, fixture.cache))

				report, err := modules.TidyWithReport(t.Context(), fixture.consumer, fixture.cache)

				require.NoError(t, err)
				assert.Empty(t, report.Imports, "an unambiguous current name does not need a historical rename")
				assert.Equal(t, tidyHistoricalConsumerProto, string(importRootsRead(t, fixture.consumer, "api/consumer/request.proto")))
				assert.Equal(t, manifest, string(importRootsRead(t, fixture.consumer, v1.ModuleFile)))
				lock := importRootsRead(t, fixture.consumer, v1.LockFile)
				assert.Equal(t, importRootsRead(t, fresh, v1.LockFile), lock)
				require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))
				assert.Equal(t, lock, importRootsRead(t, fixture.consumer, v1.LockFile))
			})
		}
	}
}

func TestTidyHistoricalCollisionsCannotGuessCurrentBinding(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"missing current name", "conflicting current copies"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			fixture := newTidyHistoricalCollisionFixture(t, false, false)
			manifest := "module example.test/consumer\nroots api\n"
			commonCommit := fixture.commits[fixture.common]
			if mode == "missing current name" {
				module := string(importRootsRead(t, fixture.common, v1.ModuleFile))
				importRootsWrite(t, fixture.common, v1.ModuleFile, strings.Replace(module, "roots proto", "roots proto/common proto/new", 1))
				commonCommit = tidyCollisionCommit(t, fixture.common)
			}
			manifest += "require " + fixture.common + " " + commonCommit + "\n"
			if mode == "conflicting current copies" {
				importRootsWrite(t, fixture.producer, "api/unrelated.proto", "syntax = \"proto3\"; package unrelated; message Added {}\n")
				producerCommit := tidyCollisionCommit(t, fixture.producer)
				manifest += "require " + fixture.producer + " " + producerCommit + "\n"
			}
			importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
			importRootsWrite(t, fixture.consumer, "api/consumer/request.proto", tidyHistoricalConsumerProto)
			lock := importRootsRead(t, fixture.consumer, v1.LockFile)

			report, err := modules.TidyWithReport(t.Context(), fixture.consumer, fixture.cache)

			require.ErrorContains(t, err, "old locked namespace has conflicting import")
			assert.Contains(t, err.Error(), "new/v1/model.proto")
			assert.Empty(t, report.Imports)
			assert.Equal(t, tidyHistoricalConsumerProto, string(importRootsRead(t, fixture.consumer, "api/consumer/request.proto")))
			assert.Equal(t, manifest, string(importRootsRead(t, fixture.consumer, v1.ModuleFile)))
			assert.Equal(t, lock, importRootsRead(t, fixture.consumer, v1.LockFile))
		})
	}
}

func newTidyHistoricalCollisionFixture(t *testing.T, conflictingRoots, reverse bool) tidyCollisionFixture {
	t.Helper()
	const previous = "syntax = \"proto3\"; package previous; message Old { int32 value = 1; }\n"
	copies := map[string]string{"api/new/v1/model.proto": previous}
	roots := ""
	if conflictingRoots {
		roots = "api extra"
		copies["extra/new/v1/model.proto"] = "syntax = \"proto3\"; package other; message Old { bool value = 1; }\n"
	}
	fixture := newTidyCollisionFixture(t, "pinned git", roots, copies, map[string]string{"proto/new/v1/model.proto": tidyHistoricalCurrentProto})
	require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache), "differing copies are initially unused")
	if reverse {
		lock, err := modules.ReadLock(filepath.Join(fixture.consumer, v1.LockFile))
		require.NoError(t, err)
		slices.Reverse(lock.Modules)
		raw, err := yaml.Marshal(lock)
		require.NoError(t, err)
		importRootsWrite(t, fixture.consumer, v1.LockFile, string(raw))
	}
	return fixture
}
