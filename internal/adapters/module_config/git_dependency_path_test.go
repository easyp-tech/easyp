package moduleconfig

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadGitDependencyAtRejectsMaterializedMetadataSymlink(t *testing.T) {
	t.Parallel()

	for _, name := range []string{dependencyManifestFile, bufWorkConfigFile, bufModuleConfigFile, bufLockFile, legacyEasyPConfigFile} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			// Reading this materialized pointer as YAML succeeds for easyp.yaml,
			// even though the historical Git mode is not a regular config file.
			require.NoError(t, os.Symlink("{}", filepath.Join(repository, name)))
			dependencyPathTestGit(t, repository, "init", "-q")
			dependencyPathTestGit(t, repository, "add", ".")
			dependencyPathTestGit(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "metadata link")
			checkout := t.TempDir()
			dependencyPathTestGit(t, checkout, "clone", "--quiet", "--no-checkout", "--", repository, checkout)
			dependencyPathTestGit(t, checkout, "-c", "core.symlinks=false", "checkout", "--quiet", "HEAD")
			info, err := os.Lstat(filepath.Join(checkout, name))
			require.NoError(t, err)
			require.True(t, info.Mode().IsRegular())

			_, err = ReadGitDependencyAt(checkout, "example.com/dependency", "")

			require.ErrorContains(t, err, "non-regular dependency config")
		})
	}
}

func TestReadGitDependencyAtRejectsModuleDirectorySymlinkBeforeMetadataRead(t *testing.T) {
	t.Parallel()

	for _, materialized := range []bool{false, true} {
		name := "symlink"
		if materialized {
			name = "materialized"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			repository, outside := t.TempDir(), t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(outside, dependencyManifestFile), []byte("module example.com/dependency\n"), 0o644))
			require.NoError(t, os.Symlink(outside, filepath.Join(repository, "api")))
			dependencyPathTestGit(t, repository, "init", "-q")
			dependencyPathTestGit(t, repository, "add", ".")
			dependencyPathTestGit(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "directory link")
			checkout := t.TempDir()
			dependencyPathTestGit(t, checkout, "clone", "--quiet", "--no-checkout", "--", repository, checkout)
			if materialized {
				dependencyPathTestGit(t, checkout, "-c", "core.symlinks=false", "checkout", "--quiet", "HEAD")
			} else {
				dependencyPathTestGit(t, checkout, "checkout", "--quiet", "HEAD")
			}

			_, err := ReadGitDependencyAt(checkout, "example.com/dependency", "api")

			require.ErrorContains(t, err, "non-regular dependency directory")
		})
	}
}

func dependencyPathTestGit(t *testing.T, directory string, args ...string) {
	t.Helper()
	command := exec.CommandContext(t.Context(), "git", args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, output)
}
