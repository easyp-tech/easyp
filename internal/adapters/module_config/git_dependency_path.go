package moduleconfig

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/adapters/gitindex"
)

// Git can materialize a symlink as a regular file with core.symlinks=false.
// Check metadata modes before reading any of its bytes. Installed snapshots
// have no Git index and retain the filesystem checks in their config reader.
func validateGitDependencyIndex(checkout string) error {
	_, err := os.Lstat(filepath.Join(checkout, ".git"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Lstat: %w", err)
	}
	command := exec.Command("git", "-C", checkout, "ls-files", "--stage", "-z")
	raw, err := command.Output()
	if err != nil {
		return fmt.Errorf("Output: %w", err)
	}
	entries, err := gitindex.Parse(string(raw))
	if err != nil {
		return fmt.Errorf("Parse: %w", err)
	}
	for _, entry := range entries {
		if IsGitDependencyConfigFile(filepath.Base(entry.Path)) && !entry.IsRegular() {
			return fmt.Errorf("non-regular dependency config %q in Git index", entry.Path)
		}
	}
	return nil
}

// Walk one bounded directory path with Lstat before any child config lookup.
// Neither a symlink nor its materialized pointer file is a module directory.
func validateGitDependencyDirectory(checkout, directory string) error {
	if directory == "" || directory == "." {
		return nil
	}
	if !filepath.IsLocal(directory) {
		return fmt.Errorf("dependency directory %q leaves the repository", directory)
	}
	current := checkout
	for _, component := range strings.Split(filepath.Clean(directory), string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Lstat: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("non-regular dependency directory %q", current)
		}
	}
	return nil
}
