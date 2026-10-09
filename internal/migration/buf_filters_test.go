package migration

import (
	"bytes"
	"testing"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/stretchr/testify/require"
)

func TestMigrationBufFiltersPreserveConsumerImports(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                             string
		excludedImport, initial, badHash bool
	}{
		{name: "historical hash includes excluded fixtures"},
		{name: "initial resolution", initial: true},
		{name: "excluded import cannot be silently dropped", excludedImport: true},
		{name: "incorrect historical hash", badHash: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				"buf.yaml":                   "version: v2\nmodules: [{path: proto, excludes: [proto/tests]}]\n",
				"proto/validate/rules.proto": "syntax = \"proto3\"; package validate; message Rules {}\n",
				"proto/tests/fixture.proto":  "syntax = \"proto3\"; package fixture; message Fixture {}\n",
			}
			repository, commit := gitSelectionRepository(t, files)
			hash := gitSelectionHash(t, map[string]string{"validate/rules.proto": files["proto/validate/rules.proto"], "tests/fixture.proto": files["proto/tests/fixture.proto"]})
			if tc.badHash {
				hash = testHash
			}
			project := t.TempDir()
			writeFixture(t, project, "easyp.yaml", "deps: ["+repository+"@"+commit+"]\ngenerate:\n  inputs: [{directory: {path: ., root: api}}]\n")
			proto := "syntax = \"proto3\"; package api; import \"validate/rules.proto\"; message Request { validate.Rules rules = 1; }\n"
			if tc.excludedImport {
				proto = "syntax = \"proto3\"; package api; import \"tests/fixture.proto\"; message Request { fixture.Fixture fixture = 1; }\n"
			}
			writeFixture(t, project, "api/request.proto", proto)
			if !tc.initial {
				writeFixture(t, project, "easyp.lock", repository+" "+commit+" "+hash+"\n")
			}
			cache := gitmodules.New(t.TempDir())
			plan, err := Build(t.Context(), Options{Dir: project, Module: "example.com/consumer", ResolveLock: true, Repository: cache})
			if tc.excludedImport {
				require.ErrorContains(t, err, "tests/fixture.proto")
				return
			}
			if tc.badHash {
				require.ErrorContains(t, err, "legacy hash mismatch")
				return
			}
			require.NoError(t, err)
			lock, err := v1.ParseLock(bytes.NewReader(outputContent(t, plan, "protobuf.lock")))
			require.NoError(t, err)
			require.Equal(t, commit, lock.Modules[0].Commit)
			require.NoError(t, cache.Install(t.Context(), lock))
			roots, err := modules.CachedSources(lock, cache)
			require.NoError(t, err)
			require.Len(t, roots, 1)
			require.NoError(t, plan.Apply())
			require.Equal(t, proto, string(mustRead(t, project, "api/request.proto")))
		})
	}
}
