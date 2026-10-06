package gitmodules

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

func TestFetchMigrationVerifiesHistoricalArchivedFileAliases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		config    string
		files     map[string]string
		link      string
		target    string
		installed map[string]string
	}{
		{
			name: "default roots preserve file pointer", files: map[string]string{"target.proto": "syntax = \"proto3\";"},
			link: "alias.proto", target: "target.proto", installed: map[string]string{"target.proto": "syntax = \"proto3\";", "alias.proto": "syntax = \"proto3\";"},
		},
		{
			name: "stripped roots preserve relative file pointer", config: "generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n",
			files: map[string]string{"proto/target.proto": "syntax = \"proto3\";"}, link: "proto/alias.proto", target: "target.proto",
			installed: map[string]string{"target.proto": "syntax = \"proto3\";", "alias.proto": "syntax = \"proto3\";"},
		},
		{
			name: "later installer rewrites cross root pointer", config: "generate:\n  inputs: [{directory: {path: proto, root: proto}}, {directory: {path: shared, root: shared}}]\n",
			files: map[string]string{"shared/target.proto": "syntax = \"proto3\";"}, link: "proto/alias.proto", target: "../shared/target.proto",
			installed: map[string]string{"target.proto": "syntax = \"proto3\";", "alias.proto": "syntax = \"proto3\";"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.config != "" {
				tt.files["easyp.yaml"] = tt.config
			}
			repository, _ := migrationTestRepository(t, tt.files)
			commit := migrationTestCommitSymlink(t, repository, tt.link, tt.target)
			oldHash := migrationTestHash(t, tt.installed)
			cacheDirectory := t.TempDir()
			fetched, err := (&Cache{root: cacheDirectory}).FetchMigration(t.Context(), repository, commit, oldHash)
			require.NoError(t, err)
			assert.Equal(t, commit, fetched.Lock.Commit)
			assert.Equal(t, commit, fetched.Lock.Version)
			assert.True(t, strings.HasPrefix(fetched.Lock.Hash, "h1:"))
			ordinary, err := (&Cache{root: t.TempDir()}).Fetch(t.Context(), repository, commit)
			require.NoError(t, err)
			assert.Equal(t, ordinary.Lock.Hash, fetched.Lock.Hash)
			assert.Equal(t, oldHash, migrationTestHash(t, tt.installed), "the recorded historical proof remains unchanged")
			migrationTestAssertNoCheckout(t, cacheDirectory)
		})
	}
}

func TestFetchMigrationInitialRootAliasDoesNotInventAnArchivedNamespace(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"easyp.yaml":      "generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n",
		"real/file.proto": "syntax = \"proto3\";",
	}
	repository, _ := migrationTestRepository(t, files)
	commit := migrationTestCommitSymlink(t, repository, "proto", "real")
	cacheDirectory := t.TempDir()
	fetched, err := (&Cache{root: cacheDirectory}).FetchMigration(t.Context(), repository, "", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"proto"}, fetched.Module.Roots)
	assert.Equal(t, commit, fetched.Lock.Commit)
	migrationTestAssertNoCheckout(t, cacheDirectory)
	// A real old archive retained real/file.proto: it never contained the
	// proto directory alias or the new file.proto import namespace.
	legacyHash := migrationTestHash(t, map[string]string{"real/file.proto": files["real/file.proto"]})
	_, err = (&Cache{root: cacheDirectory}).FetchMigration(t.Context(), repository, commit, legacyHash)
	require.ErrorContains(t, err, "source selection")
	require.ErrorContains(t, err, "manual migration")
	migrationTestAssertNoCheckout(t, cacheDirectory)
}

func TestFetchMigrationRelevantAliasStillRequiresCompleteSourceOwnership(t *testing.T) {
	t.Parallel()
	for _, link := range []string{"root", "metadata", "file"} {
		t.Run(link, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				"proto/file.proto":   "syntax = \"proto3\";",
				"extras/extra.proto": "unrelated source scope",
				"easyp.yaml":         "generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n",
			}
			name, target := "proto/alias.proto", "file.proto"
			if link == "root" {
				files["easyp.yaml"] = "generate:\n  inputs: [{directory: {path: api, root: api}}]\n"
				name, target = "api", "proto"
			}
			if link == "metadata" {
				files["config.yaml"] = files["easyp.yaml"]
				delete(files, "easyp.yaml")
				name, target = "easyp.yaml", "config.yaml"
			}
			repository, _ := migrationTestRepository(t, files)
			migrationTestCommitSymlink(t, repository, name, target)
			cacheDirectory := t.TempDir()
			_, err := (&Cache{root: cacheDirectory}).FetchMigration(t.Context(), repository, "", "")
			require.ErrorContains(t, err, "source selection")
			require.ErrorContains(t, err, "extras/extra.proto")
			migrationTestAssertNoCheckout(t, cacheDirectory)
		})
	}
}

func TestMigrationArchiveCandidatesPreserveBothEvidencedPointerPolicies(t *testing.T) {
	t.Parallel()
	repository, _ := migrationTestRepository(t, map[string]string{"shared/target.proto": "proto"})
	commit := migrationTestCommitSymlink(t, repository, "proto/alias.proto", "../shared/target.proto")
	tracked, err := migrationTrackedFiles(t.Context(), repository)
	require.NoError(t, err)
	nodes, err := readMigrationProtoArchive(t.Context(), repository, commit, tracked.trackedFiles)
	require.NoError(t, err)
	roots := []string{"proto", "shared"}
	v15, _, err := installMigrationArchive(nodes, roots, false)
	require.NoError(t, err)
	assert.Equal(t, "../shared/target.proto", string(v15["alias.proto"].data))
	_, err = sourceview.New(v15).Open(t.Context(), "alias.proto")
	require.ErrorIs(t, err, sourceview.ErrOutsideRoot)
	v16, _, err := installMigrationArchive(nodes, roots, true)
	require.NoError(t, err)
	assert.Equal(t, "target.proto", string(v16["alias.proto"].data))
	selected := map[string][]byte{"alias.proto": []byte("proto"), "target.proto": []byte("proto")}
	hashes, err := migrationArchiveHashes(t.Context(), nodes, roots, selected)
	require.NoError(t, err)
	assert.Equal(t, []string{migrationTestHash(t, map[string]string{"alias.proto": "proto", "target.proto": "proto"})}, hashes)
}

func TestMigrationArchiveProofRejectsInvalidPointerLeaves(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		target string
		want   error
	}{
		{name: "external", target: "/outside/target.proto", want: sourceview.ErrOutsideRoot},
		{name: "drive", target: "C:/outside", want: sourceview.ErrUnsupported},
		{name: "NUL", target: "bad\x00target", want: sourceview.ErrUnsupported},
		{name: "dangling", target: "missing.proto", want: fs.ErrNotExist},
		{name: "cycle", target: "alias.proto", want: sourceview.ErrCycle},
		{name: "directory leaf", target: "directory", want: sourceview.ErrUnsupported},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			nodes := []migrationArchiveNode{
				{name: "directory/", mode: fs.ModeDir},
				{name: "directory/target.proto", data: []byte("proto")},
				{name: "alias.proto", mode: fs.ModeSymlink, data: []byte(tt.target)},
			}
			_, err := migrationArchiveHashes(t.Context(), nodes, nil, map[string][]byte{"alias.proto": []byte("proto"), "directory/target.proto": []byte("proto")})
			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestFetchMigrationNativeInitialAliasesStillRejectImportCollisions(t *testing.T) {
	t.Parallel()
	repository, _ := migrationTestRepository(t, map[string]string{"real/file.proto": "proto"})
	migrationTestWriteFiles(t, repository, map[string]string{"protobuf.mod": "module " + repository + "\nroots one two\n"})
	migrationTestCommitSymlink(t, repository, "one", "real")
	migrationTestCommitSymlink(t, repository, "two", "real")
	cacheDirectory := t.TempDir()
	_, err := (&Cache{root: cacheDirectory}).FetchMigration(t.Context(), repository, "", "")
	require.ErrorContains(t, err, "source selection collides")
	require.ErrorContains(t, err, "file.proto")
	migrationTestAssertNoCheckout(t, cacheDirectory)
}

func TestFetchMigrationArchiveCannotBorrowOmittedAliasTarget(t *testing.T) {
	t.Parallel()
	files := map[string]string{"target.txt": "unarchived target"}
	repository, _ := migrationTestRepository(t, files)
	commit := migrationTestCommitSymlink(t, repository, "alias.proto", "target.txt")
	legacyHash := migrationTestHash(t, map[string]string{"alias.proto": files["target.txt"]})
	cacheDirectory := t.TempDir()
	_, err := (&Cache{root: cacheDirectory}).FetchMigration(t.Context(), repository, commit, legacyHash)
	require.ErrorIs(t, err, fs.ErrNotExist)
	assert.ErrorContains(t, err, "migrationArchiveHashes")
	migrationTestAssertNoCheckout(t, cacheDirectory)
}

func TestFetchMigrationArchiveAliasReadsArchivedTargetBytes(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		".gitattributes": "target.proto export-subst\n",
		"target.proto":   "// $Format:%H$\nsyntax = \"proto3\";",
	}
	repository, _ := migrationTestRepository(t, files)
	commit := migrationTestCommitSymlink(t, repository, "alias.proto", "target.proto")
	archived := strings.ReplaceAll(files["target.proto"], "$Format:%H$", commit)
	legacyHash := migrationTestHash(t, map[string]string{"alias.proto": archived, "target.proto": archived})
	cacheDirectory := t.TempDir()
	_, err := (&Cache{root: cacheDirectory}).FetchMigration(t.Context(), repository, commit, legacyHash)
	require.ErrorContains(t, err, "source selection")
	require.ErrorContains(t, err, "manual migration")
	migrationTestAssertNoCheckout(t, cacheDirectory)
}
