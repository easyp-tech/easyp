package gitmodules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchMigrationMatchingWholeTreePinBypassesProtoArchive(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		".gitattributes": "file.proto export-ignore\n",
		"file.proto":     "syntax = \"proto3\";\n",
		"README":         "part of the historical whole-tree representation\n",
	}
	repository, commit := migrationTestRepository(t, files)
	expectedHash := migrationTestHash(t, files)

	fetched, err := (&Cache{root: t.TempDir()}).FetchMigration(t.Context(), repository, commit, expectedHash)

	require.NoError(t, err)
	assert.Equal(t, commit, fetched.Lock.Commit)
	assert.Equal(t, snapshotTestHash(t, files), fetched.Lock.Hash)
	assert.NotEqual(t, expectedHash, fetched.Lock.Hash, "historical proof and native projection have separate scopes")
}

func TestFetchMigrationInitialResolutionIgnoresUnrelatedFileCollision(t *testing.T) {
	t.Parallel()

	repository, _ := migrationTestRepository(t, map[string]string{
		"easyp.yaml":       "generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n",
		"proto/file.proto": "syntax = \"proto3\";\n",
		"proto/README":     "inside root\n",
		"README":           "outside root\n",
	})
	cacheDir := t.TempDir()

	fetched, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, "", "")

	require.NoError(t, err)
	assert.Equal(t, repository, fetched.Lock.Source)
	migrationTestAssertNoCheckout(t, cacheDir)
}

func TestFetchMigrationHashMismatchListsOnlyEligibleProofs(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name          string
		auxiliaryLink bool
	}{
		{name: "regular tree permits both proofs"},
		{name: "omitted links require archive proof", auxiliaryLink: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				"file.proto": "syntax = \"proto3\";\n",
				"README":     "retained regular content\n",
			}
			repository, commit := migrationTestRepository(t, files)
			if tt.auxiliaryLink {
				commit = migrationTestCommitSymlink(t, repository, ".bazelrc", "file.proto")
			}
			wrongHash := migrationTestHash(t, map[string]string{"different": "contents"})
			cacheDir := t.TempDir()

			fetched, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, commit, wrongHash)

			require.ErrorContains(t, err, "legacy hash mismatch")
			archiveHash := migrationTestHash(t, map[string]string{"file.proto": files["file.proto"]})
			assert.ErrorContains(t, err, "got archive "+archiveHash)
			if tt.auxiliaryLink {
				assert.NotContains(t, err.Error(), "whole-tree")
			} else {
				assert.ErrorContains(t, err, "or whole-tree "+migrationTestHash(t, files))
			}
			assert.Empty(t, fetched.Lock.Source)
			migrationTestAssertNoCheckout(t, cacheDir)
		})
	}
}
