package gitmodules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchMigrationWithRootsSeparatesHistoricalAndInputNamespaces(t *testing.T) {
	t.Parallel()
	files := map[string]string{"api/service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n"}
	repository, commit := migrationTestRepository(t, files)
	cacheDirectory := t.TempDir()
	fetched, err := (&Cache{root: cacheDirectory}).FetchMigrationWithRoots(t.Context(), repository, commit, migrationTestHash(t, files), []string{"api"})
	require.NoError(t, err)
	assert.Equal(t, []string{"api"}, fetched.Module.Roots)
	assert.Equal(t, []string{"api"}, fetched.Lock.Roots)
	require.NotNil(t, fetched.Inspection)
	assert.Equal(t, map[string]string{"api/service/a.proto": "api/service/a.proto"}, fetched.Inspection.LegacyFiles)
	require.Len(t, fetched.Inspection.Files, 1)
	assert.Equal(t, files["api/service/a.proto"], string(fetched.Inspection.Files[0].Content))
	migrationTestAssertNoCheckout(t, cacheDirectory)
}

func TestFetchMigrationWithRootsRetainsArchiveByteEquivalenceChecks(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, attribute, body string }{
		{name: "export ignored source", attribute: "api/changed.proto export-ignore\n", body: "syntax = \"proto3\"; package changed.v1; message Changed {}\n"},
		{name: "export substituted source", attribute: "api/changed.proto export-subst\n", body: "// $Format:%H$\nsyntax = \"proto3\"; package changed.v1; message Changed {}\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{".gitattributes": tt.attribute, "api/kept.proto": "syntax = \"proto3\"; package kept.v1; message Kept {}\n", "api/changed.proto": tt.body}
			repository, commit := migrationTestRepository(t, files)
			installed := map[string]string{"api/kept.proto": files["api/kept.proto"]}
			if tt.name == "export substituted source" {
				installed["api/changed.proto"] = strings.ReplaceAll(tt.body, "$Format:%H$", commit)
			}
			cacheDirectory := t.TempDir()
			fetched, err := (&Cache{root: cacheDirectory}).FetchMigrationWithRoots(t.Context(), repository, commit, migrationTestHash(t, installed), []string{"api"})
			require.ErrorContains(t, err, "source selection")
			assert.ErrorContains(t, err, "api/changed.proto")
			assert.Empty(t, fetched.Lock.Source)
			migrationTestAssertNoCheckout(t, cacheDirectory)
		})
	}
}

func TestFetchMigrationWithRootsReturnsVerifiedArchiveMapping(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"easyp.yaml":          "generate:\n  inputs: [{directory: {path: ., root: api}}]\n",
		"api/service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n",
	}
	repository, commit := migrationTestRepository(t, files)
	installed := map[string]string{"service/a.proto": files["api/service/a.proto"]}
	fetched, err := (&Cache{root: t.TempDir()}).FetchMigrationWithRoots(t.Context(), repository, commit, migrationTestHash(t, installed), []string{})
	require.NoError(t, err)
	require.NotNil(t, fetched.Inspection)
	assert.Equal(t, map[string]string{"service/a.proto": "api/service/a.proto"}, fetched.Inspection.LegacyFiles)
	assert.Equal(t, []string{"api"}, fetched.Module.Roots)
	assert.Empty(t, fetched.Lock.Roots)
}

func TestFetchMigrationWithRootsChecksHashBeforeRootAuthority(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"easyp.yaml":    "generate:\n  inputs: [{directory: {path: ., root: proto}}]\n",
		"proto/a.proto": "syntax = \"proto3\"; package a.v1; message A {}\n",
	}
	repository, commit := migrationTestRepository(t, files)
	_, err := (&Cache{root: t.TempDir()}).FetchMigrationWithRoots(t.Context(), repository, commit, migrationTestHash(t, map[string]string{"wrong.proto": "wrong"}), []string{"api"})
	require.ErrorContains(t, err, "legacy hash mismatch")
	assert.NotContains(t, err.Error(), "authoritative roots")
}
