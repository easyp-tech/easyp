package modules_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestTidyBufExcludesRemoveConflictingAndMalformedCandidates(t *testing.T) {
	t.Parallel()
	for _, metadata := range []struct {
		name   string
		prefix string
		files  map[string]string
	}{
		{name: "v1", files: map[string]string{"buf.yaml": "version: v1\nbuild:\n  excludes: [common, preview]\n"}},
		{name: "v2", prefix: "api/", files: map[string]string{"buf.yaml": "version: v2\nmodules:\n  - path: api\n    excludes: [api/common, api/preview]\n"}},
		{name: "v1_workspace", prefix: "api/", files: map[string]string{
			"buf.work.yaml": "version: v1\ndirectories: [api]\n",
			"api/buf.yaml":  "version: v1\nbuild:\n  excludes: [common, preview]\n",
		}},
	} {
		for _, mode := range []string{"pinned git", "local replace"} {
			t.Run(metadata.name+"/"+mode, func(t *testing.T) {
				t.Parallel()
				const malformed = "this excluded source must not be parsed"
				fixture := newTidyCollisionFixture(t, mode, "", map[string]string{
					"api/common/v1/types.proto":    strings.Replace(tidyCommonProto, "string value", "bool value", 1),
					"api/preview/broken.proto":     malformed,
					"api/previewed/retained.proto": "syntax = \"proto3\"; package retained; message Retained {}\n",
				}, nil)
				require.NoError(t, os.Remove(filepath.Join(fixture.producer, v1.ModuleFile)))
				if metadata.prefix == "" {
					entries, err := os.ReadDir(filepath.Join(fixture.producer, "api"))
					require.NoError(t, err)
					for _, entry := range entries {
						require.NoError(t, os.Rename(filepath.Join(fixture.producer, "api", entry.Name()), filepath.Join(fixture.producer, entry.Name())))
					}
					require.NoError(t, os.Remove(filepath.Join(fixture.producer, "api")))
				}
				for name, content := range metadata.files {
					importRootsWrite(t, fixture.producer, name, content)
				}
				commit := tidyCollisionCommit(t, fixture.producer)
				manifest := strings.Replace(string(importRootsRead(t, fixture.consumer, v1.ModuleFile)), fixture.commits[fixture.producer], commit, 1)
				importRootsWrite(t, fixture.consumer, v1.ModuleFile, manifest)
				const localLock = "version: 1\nmodules: []\n# local developer lock\n"
				if mode == "local replace" {
					importRootsWrite(t, fixture.consumer, v1.LockFile, localLock)
				}

				require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))

				assert.NoFileExists(t, filepath.Join(fixture.producer, v1.ModuleFile))
				assert.Equal(t, tidyCommonProto, string(importRootsRead(t, fixture.common, "proto/common/v1/types.proto")))
				assert.Equal(t, strings.Replace(tidyCommonProto, "string value", "bool value", 1), string(importRootsRead(t, fixture.producer, metadata.prefix+"common/v1/types.proto")))
				assert.Equal(t, malformed, string(importRootsRead(t, fixture.producer, metadata.prefix+"preview/broken.proto")))
				for name, content := range metadata.files {
					assert.Equal(t, content, string(importRootsRead(t, fixture.producer, name)))
				}
				if mode == "local replace" {
					assert.Equal(t, manifest, string(importRootsRead(t, fixture.consumer, v1.ModuleFile)))
					assert.Equal(t, localLock, string(importRootsRead(t, fixture.consumer, v1.LockFile)))
				}
				beforeManifest := importRootsRead(t, fixture.consumer, v1.ModuleFile)
				beforeLock := importRootsRead(t, fixture.consumer, v1.LockFile)
				require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))
				assert.Equal(t, beforeManifest, importRootsRead(t, fixture.consumer, v1.ModuleFile))
				assert.Equal(t, beforeLock, importRootsRead(t, fixture.consumer, v1.LockFile))
			})
		}
	}
}
