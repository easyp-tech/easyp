package moduleconfig

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	for _, entry := range strings.Split(string(raw), "\x00") {
		if entry == "" {
			continue
		}
		metadata, name, ok := strings.Cut(entry, "\t")
		fields := strings.Fields(metadata)
		if !ok || len(fields) != 3 || fields[2] != "0" {
			return fmt.Errorf("invalid Git index entry %q", entry)
		}
		if !filepath.IsLocal(filepath.FromSlash(name)) {
			return fmt.Errorf("invalid tracked file path %q", name)
		}
		if IsGitDependencyConfigFile(filepath.Base(name)) && fields[0] != "100644" && fields[0] != "100755" {
			return fmt.Errorf("non-regular dependency config %q in Git index", name)
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
