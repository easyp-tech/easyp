package migration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestMigrationWithoutLockKeepsPinnedGitAndFilteredImports(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name       string
		transitive bool
	}{
		{name: "direct pinned provider"},
		{name: "provider of pinned Git generation input", transitive: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			excluded := "proto/protovalidate-testing/tests/example/v1/example.proto"
			files := map[string]string{
				"buf.yaml": "version: v2\nmodules:\n  - path: proto/protovalidate\n  - path: proto/protovalidate-testing\n    excludes: [proto/protovalidate-testing/tests]\n",
				"proto/protovalidate/buf/validate/validate.proto":                 "syntax = \"proto3\"; package buf.validate; message Rule { string value = 1; }\n",
				"proto/protovalidate-testing/buf/validate/conformance/case.proto": "syntax = \"proto3\"; package buf.validate.conformance; message Case {}\n",
				excluded: "syntax = \"proto3\"; package example.v1; message Example {}\n",
			}
			provider, pinned := gitSelectionRepository(t, files)
			// HEAD no longer has the reported file or the imported contract. The
			// explicit SHA must still own both historical proof and source bytes.
			require.NoError(t, os.Remove(filepath.Join(provider, filepath.FromSlash(excluded))))
			writeFixture(t, provider, "proto/protovalidate/buf/validate/validate.proto", "syntax = \"proto3\"; package buf.validate; message Changed {}\n")
			gitSelectionCommand(t, provider, "add", "-A")
			gitSelectionCommand(t, provider, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "newer incompatible revision")

			consumer := t.TempDir()
			local := "syntax = \"proto3\"; package consumer.v1; import \"buf/validate/validate.proto\"; message Request { buf.validate.Rule rule = 1; }\n"
			config := "deps: [" + provider + "@" + pinned + "]\ngenerate:\n  inputs: [{directory: {path: ., root: api}}]\n  plugins: [{name: python, out: gen/python}]\n"
			var primary, primaryPin string
			if tt.transitive {
				primary, primaryPin = gitSelectionRepository(t, map[string]string{
					"easyp.yaml":                 "deps: [" + provider + "@" + pinned + "]\ngenerate:\n  inputs: [{directory: {path: ., root: api}}]\n",
					"api/remote/v1/remote.proto": "syntax = \"proto3\"; package remote.v1; import \"buf/validate/validate.proto\"; message Remote { buf.validate.Rule rule = 1; }\n",
				})
				config = "generate:\n  inputs:\n    - directory: {path: ., root: api}\n    - git_repo: {url: " + primary + "@" + primaryPin + "}\n  plugins: [{name: python, out: gen/python}]\n"
				local = "syntax = \"proto3\"; package consumer.v1; import \"remote/v1/remote.proto\"; message Request { remote.v1.Remote remote = 1; }\n"
			}
			writeFixture(t, consumer, "easyp.yaml", config)
			writeFixture(t, consumer, "api/request.proto", local)
			require.NoFileExists(t, filepath.Join(consumer, "easyp.lock"))
			cache := gitmodules.New(t.TempDir())

			plan, err := Build(t.Context(), Options{Dir: consumer, Module: "example.test/consumer", ResolveLock: true, Repository: cache})

			require.NoError(t, err)
			lock, err := v1.ParseLock(bytes.NewReader(outputContent(t, plan, v1.LockFile)))
			require.NoError(t, err)
			pins := make(map[string]v1.LockedModule)
			for _, entry := range lock.Modules {
				pins[entry.Source] = entry
			}
			require.Contains(t, pins, provider)
			require.Equal(t, pinned, pins[provider].Commit)
			require.Equal(t, pinned, pins[provider].Version)
			if tt.transitive {
				require.Len(t, pins, 2)
				require.Equal(t, primaryPin, pins[primary].Commit)
			}
			require.NoError(t, cache.Install(t.Context(), lock))
			installed, _, err := cache.Cached(pins[provider])
			require.NoError(t, err)
			require.NoFileExists(t, filepath.Join(installed, filepath.FromSlash(excluded)))
			raw, err := os.ReadFile(filepath.Join(installed, "proto/protovalidate/buf/validate/validate.proto"))
			require.NoError(t, err)
			require.Equal(t, files["proto/protovalidate/buf/validate/validate.proto"], string(raw))
			require.NoError(t, plan.Apply())
			require.NoFileExists(t, filepath.Join(consumer, "easyp.lock"))
			require.Equal(t, local, string(mustRead(t, consumer, "api/request.proto")))
			require.Equal(t, config, string(mustRead(t, consumer, "easyp.yaml.v0.bak")))
		})
	}
}
