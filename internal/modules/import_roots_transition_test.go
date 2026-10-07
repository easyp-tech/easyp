package modules_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestUpdateRequiresExplicitAdoptionOfNewAuthoritativeNamespace(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		root        string
		newImport   string
		declaration string
	}{
		{name: "declared_root", root: "api", newImport: "service/v1/svc.proto", declaration: "roots api\n"},
		{name: "authoritative_default", root: ".", newImport: "api/service/v1/svc.proto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, oldCommit := importRootsRepository(t, map[string]string{
				"api/service/v1/a.proto":   `syntax = "proto3"; import "svc.proto";`,
				"api/service/v1/svc.proto": `syntax = "proto3";`,
			}, nil)
			importRootsGit(t, repository, "tag", "v0.4.0")
			root := importRootsConsumer(t, repository, "v0.4.0", `syntax = "proto3"; import "svc.proto";`)
			cache := gitmodules.New(t.TempDir())
			require.NoError(t, modules.Tidy(t.Context(), root, cache))
			old, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Equal(t, []string{"api/service/v1"}, old.Modules[0].Roots)
			importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\n"+tt.declaration)
			importRootsWrite(t, repository, "api/service/v1/a.proto", "syntax = \"proto3\"; import \""+tt.newImport+"\";")
			importRootsGit(t, repository, "add", ".")
			importRootsGit(t, repository, "commit", "--quiet", "-m", "declare a new namespace")
			importRootsGit(t, repository, "tag", "v0.5.0")
			newCommit := importRootsGit(t, repository, "rev-parse", "HEAD")
			// Even a consumer already adjusted for the new names does not
			// authorize an implicit update to replace a verified namespace.
			importRootsWrite(t, root, "consumer.proto", "syntax = \"proto3\"; import \""+tt.newImport+"\";")
			manifest, lock := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)

			spy := &importRootsInstallSpy{Cache: cache}
			err = modules.Update(t.Context(), root, spy)

			require.ErrorContains(t, err, "import namespace")
			assert.ErrorContains(t, err, repository)
			assert.ErrorContains(t, err, "v0.4.0")
			assert.ErrorContains(t, err, oldCommit)
			assert.ErrorContains(t, err, "v0.5.0")
			assert.ErrorContains(t, err, newCommit)
			assert.ErrorContains(t, err, "api/service/v1")
			assert.ErrorContains(t, err, "authoritative dependency metadata")
			assert.ErrorContains(t, err, "svc.proto -> "+tt.newImport)
			assert.ErrorContains(t, err, "generated SDK paths")
			assert.ErrorContains(t, err, "easyp get --import-root '"+tt.root+"'")
			assert.ErrorContains(t, err, repository+"@v0.5.0")
			assert.ErrorContains(t, err, repository+"@"+oldCommit)
			assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
			assert.Equal(t, lock, importRootsRead(t, root, v1.LockFile))
			assert.Zero(t, spy.installs)

			importRootsWrite(t, root, "consumer.proto", `syntax = "proto3"; import "svc.proto";`)
			err = modules.GetWithRoots(t.Context(), root, v1.Requirement{Module: repository, Version: "v0.5.0"}, cache, []string{tt.root})
			require.ErrorContains(t, err, "cannot resolve imports")
			assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
			assert.Equal(t, lock, importRootsRead(t, root, v1.LockFile))
			importRootsWrite(t, root, "consumer.proto", "syntax = \"proto3\"; import \""+tt.newImport+"\";")

			err = modules.GetWithRoots(t.Context(), root, v1.Requirement{Module: repository, Version: "v0.5.0"}, cache, []string{tt.root})

			require.NoError(t, err)
			adopted, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Equal(t, "v0.5.0", adopted.Modules[0].Version)
			assert.Equal(t, newCommit, adopted.Modules[0].Commit)
			assert.Empty(t, adopted.Modules[0].Roots)
		})
	}
}

func TestNewAuthoritativeRootsKeepExistingDescriptorNames(t *testing.T) {
	t.Parallel()
	repository, _ := importRootsRepository(t, map[string]string{"proto/a.proto": `syntax = "proto3"; import "svc.proto";`, "proto/svc.proto": `syntax = "proto3";`}, nil)
	importRootsGit(t, repository, "tag", "v0.4.0")
	root := importRootsConsumer(t, repository, "v0.4.0", `syntax = "proto3"; import "svc.proto";`)
	cache := gitmodules.New(t.TempDir())
	require.NoError(t, modules.Tidy(t.Context(), root, cache))
	importRootsGit(t, repository, "mv", "proto", "api")
	importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\nroots api\n")
	importRootsGit(t, repository, "add", ".")
	importRootsGit(t, repository, "commit", "--quiet", "-m", "move files without renaming descriptors")
	importRootsGit(t, repository, "tag", "v0.5.0")

	err := modules.Update(t.Context(), root, cache)

	require.NoError(t, err)
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	assert.Equal(t, "v0.5.0", lock.Modules[0].Version)
	assert.Empty(t, lock.Modules[0].Roots)
	sources, err := modules.CachedSources(lock, cache)
	require.NoError(t, err)
	owners, err := sources.FileModules()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"a.proto": repository, "svc.proto": repository}, owners)
}

func TestNewAuthorityChecksOmittedDefaultFallback(t *testing.T) {
	t.Parallel()
	repository, oldCommit := importRootsRepository(t, map[string]string{"api/svc.proto": `syntax = "proto3";`}, nil)
	importRootsGit(t, repository, "tag", "v0.4.0")
	root := importRootsConsumer(t, repository, "v0.4.0", `syntax = "proto3"; import "api/svc.proto";`)
	cache := gitmodules.New(t.TempDir())
	require.NoError(t, modules.Tidy(t.Context(), root, cache))
	old, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	assert.Empty(t, old.Modules[0].Roots)
	importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\nroots api\n")
	importRootsGit(t, repository, "add", ".")
	importRootsGit(t, repository, "commit", "--quiet", "-m", "declare roots for previously default namespace")
	importRootsGit(t, repository, "tag", "v0.5.0")
	newCommit := importRootsGit(t, repository, "rev-parse", "HEAD")
	importRootsWrite(t, root, "consumer.proto", `syntax = "proto3"; import "svc.proto";`)
	manifest, before := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)
	spy := &importRootsInstallSpy{Cache: cache}

	err = modules.Update(t.Context(), root, spy)

	require.ErrorContains(t, err, "import namespace")
	assert.ErrorContains(t, err, "verified fallback roots [.]")
	assert.ErrorContains(t, err, "api/svc.proto -> svc.proto")
	assert.ErrorContains(t, err, oldCommit)
	assert.ErrorContains(t, err, newCommit)
	assert.Zero(t, spy.installs)
	assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
	assert.Equal(t, before, importRootsRead(t, root, v1.LockFile))
	require.NoError(t, modules.GetWithRoots(t.Context(), root, v1.Requirement{Module: repository, Version: "v0.5.0"}, cache, []string{"api"}))
}

func TestRootTransitionKeepsPriorAuthorityAndIdenticalEffectiveRoots(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oldFile     string
		oldContent  string
		newRoots    string
		newConsumer string
		moveFile    bool
	}{
		{name: "prior_native_default", oldFile: v1.ModuleFile, oldContent: "module %s\n", newRoots: "roots api\n", newConsumer: `syntax = "proto3"; import "svc.proto";`},
		{name: "prior_buf_default", oldFile: "buf.yaml", oldContent: "version: v2\nmodules: [{path: .}]\n", newRoots: "roots api\n", newConsumer: `syntax = "proto3"; import "svc.proto";`},
		{name: "same_effective_default_allows_physical_change", newConsumer: `syntax = "proto3"; import "api/other.proto";`, moveFile: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, _ := importRootsRepository(t, map[string]string{"api/svc.proto": `syntax = "proto3";`}, nil)
			if tt.oldFile != "" {
				content := strings.ReplaceAll(tt.oldContent, "%s", repository)
				importRootsWrite(t, repository, tt.oldFile, content)
				importRootsGit(t, repository, "add", ".")
				importRootsGit(t, repository, "commit", "--quiet", "-m", "old authority")
			}
			importRootsGit(t, repository, "tag", "v0.4.0")
			root := importRootsConsumer(t, repository, "v0.4.0", `syntax = "proto3"; import "api/svc.proto";`)
			cache := gitmodules.New(t.TempDir())
			require.NoError(t, modules.Tidy(t.Context(), root, cache))
			old, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Empty(t, old.Modules[0].Roots)
			if tt.moveFile {
				importRootsGit(t, repository, "mv", "api/svc.proto", "api/other.proto")
			}
			importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\n"+tt.newRoots)
			importRootsGit(t, repository, "add", ".")
			importRootsGit(t, repository, "commit", "--quiet", "-m", "new native metadata")
			importRootsGit(t, repository, "tag", "v0.5.0")
			importRootsWrite(t, root, "consumer.proto", tt.newConsumer)

			err = modules.Update(t.Context(), root, cache)

			require.NoError(t, err)
			updated, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Equal(t, "v0.5.0", updated.Modules[0].Version)
		})
	}
}

func TestNewAuthorityRejectsShadowedOldSourceBinding(t *testing.T) {
	t.Parallel()
	repository, oldCommit := importRootsRepository(t, map[string]string{"api/v1/svc.proto": `syntax = "proto3";`}, nil)
	importRootsGit(t, repository, "tag", "v0.4.0")
	root := importRootsConsumer(t, repository, "v0.4.0", `syntax = "proto3"; import "svc.proto";`)
	cache := gitmodules.New(t.TempDir())
	require.NoError(t, modules.GetWithRoots(t.Context(), root, v1.Requirement{Module: repository, Version: "v0.4.0"}, cache, []string{"api/v1"}))
	importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\nroots api\n")
	importRootsWrite(t, repository, "api/svc.proto", `syntax = "proto3"; package newer;`)
	importRootsGit(t, repository, "add", ".")
	importRootsGit(t, repository, "commit", "--quiet", "-m", "rename old source and reuse its import name")
	importRootsGit(t, repository, "tag", "v0.5.0")
	newCommit := importRootsGit(t, repository, "rev-parse", "HEAD")
	manifest, before := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)
	spy := &importRootsInstallSpy{Cache: cache}

	err := modules.Update(t.Context(), root, spy)

	require.ErrorContains(t, err, "import namespace")
	assert.ErrorContains(t, err, "svc.proto -> v1/svc.proto")
	assert.ErrorContains(t, err, "verified fallback roots [api/v1]")
	assert.ErrorContains(t, err, "roots [api] from authoritative dependency metadata")
	assert.ErrorContains(t, err, oldCommit)
	assert.ErrorContains(t, err, newCommit)
	assert.Zero(t, spy.installs)
	assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
	assert.Equal(t, before, importRootsRead(t, root, v1.LockFile))
}
