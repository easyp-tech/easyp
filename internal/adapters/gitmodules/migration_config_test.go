package gitmodules

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMigrationBufRootsAcceptsSourceFilters(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, fields string }{
		{name: "excludes", fields: "excludes: [private]"},
		{name: "includes", fields: "includes: [public]"},
		{name: "both", fields: "includes: [public], excludes: [public/private]"},
		{name: "empty", fields: "excludes: []"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw := []byte("version: v2\nmodules: [{path: proto, " + tc.fields + "}]\n")
			roots, err := parseMigrationLegacyRoots("buf.yaml", raw)
			require.NoError(t, err)
			// Filtering was not part of the v0 archive rename algorithm.
			require.Equal(t, []string{"proto"}, roots)
		})
	}
}

func TestFetchMigrationPreservesBufExcludesArchiveHash(t *testing.T) {
	t.Parallel()
	config := "version: v2\nmodules: [{path: proto/public}, {path: proto/testing, excludes: [proto/testing/tests]}]\n"
	files := map[string]string{
		"buf.yaml":                        config,
		"proto/public/api.proto":          "syntax = \"proto3\"; package api; message API {}\n",
		"proto/testing/conformance.proto": "syntax = \"proto3\"; package tests; message Conformance {}\n",
		"proto/testing/tests/data.txt":    "excluded non-proto test fixture\n",
	}
	repository, commit := migrationTestRepository(t, files)
	hash := migrationTestHash(t, map[string]string{
		"api.proto":         files["proto/public/api.proto"],
		"conformance.proto": files["proto/testing/conformance.proto"],
	})
	cache := New(t.TempDir())
	fetched, err := cache.FetchMigration(t.Context(), repository, commit, hash)
	require.NoError(t, err)
	require.Equal(t, commit, fetched.Lock.Commit)
	require.NotEmpty(t, fetched.Module.ProtoFilters)
	require.Equal(t, "proto/public/api.proto", fetched.Inspection.LegacyFiles["api.proto"])
	require.Equal(t, "proto/testing/conformance.proto", fetched.Inspection.LegacyFiles["conformance.proto"])
	require.NotContains(t, fetched.Inspection.LegacyFiles, "tests/data.txt")
}
