package go_git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/easyp-tech/easyp/internal/core"
)

// Snapshot contains revision-local protobuf sources and dependency metadata.
// It never checks out or modifies the caller's working tree.
type Snapshot struct {
	Root           string
	RepositoryRoot string
}

// RepositoryRoot returns the current worktree root without resolving a
// baseline revision or modifying the checkout.
func RepositoryRoot(directory string) (string, error) {
	repository, err := gogit.PlainOpenWithOptions(directory, &gogit.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	if errors.Is(err, gogit.ErrRepositoryNotExists) {
		return "", core.ErrRepositoryDoesNotExist
	}
	if err != nil {
		return "", fmt.Errorf("PlainOpenWithOptions: %w", err)
	}
	worktree, err := repository.Worktree()
	if err != nil {
		return "", fmt.Errorf("Worktree: %w", err)
	}
	return worktree.Filesystem.Root(), nil
}

// Close removes the temporary snapshot.
func (s *Snapshot) Close() error { return os.RemoveAll(s.Root) }

// SnapshotRevision materializes a revision's protobuf inputs, including hidden
// local replacement directories. Only relevant regular files are materialized.
func SnapshotRevision(ctx context.Context, directory, ref string) (_ *Snapshot, err error) {
	repository, err := gogit.PlainOpenWithOptions(directory, &gogit.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	if errors.Is(err, gogit.ErrRepositoryNotExists) {
		return nil, core.ErrRepositoryDoesNotExist
	}
	if err != nil {
		return nil, fmt.Errorf("PlainOpenWithOptions: %w", err)
	}
	wt, err := repository.Worktree()
	if err != nil {
		return nil, fmt.Errorf("Worktree: %w", err)
	}
	commit, err := baselineCommit(repository, ref)
	if err != nil {
		return nil, fmt.Errorf("baselineCommit: %w", err)
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("Tree: %w", err)
	}
	root, err := os.MkdirTemp("", "easyp-breaking-*")
	if err != nil {
		return nil, fmt.Errorf("MkdirTemp: %w", err)
	}
	snapshot := &Snapshot{Root: root, RepositoryRoot: wt.Filesystem.Root()}
	defer func() {
		if err != nil {
			_ = snapshot.Close()
		}
	}()
	err = tree.Files().ForEach(func(file *object.File) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !snapshotInput(file.Name) {
			return nil
		}
		relative := filepath.FromSlash(file.Name)
		if !filepath.IsLocal(relative) {
			return fmt.Errorf("invalid snapshot path %q", file.Name)
		}
		if file.Mode != filemode.Regular && file.Mode != filemode.Executable && file.Mode != filemode.Deprecated {
			// Additional policy candidates are copied only when regular. An unused
			// YAML symlink must not break an otherwise unrelated baseline check. If
			// referenced as a policy, its absence is reported without following it.
			if snapshotAdditionalPolicy(file.Name) {
				return nil
			}
			return fmt.Errorf("baseline input %q is not a regular file", file.Name)
		}
		reader, err := file.Reader()
		if err != nil {
			return fmt.Errorf("Reader: %w", err)
		}
		raw, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return fmt.Errorf("ReadAll: %w", readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("Close: %w", closeErr)
		}
		target := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return fmt.Errorf("MkdirAll: %w", err)
		}
		if err := os.WriteFile(target, raw, 0o600); err != nil {
			return fmt.Errorf("WriteFile: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("ForEach: %w", err)
	}
	return snapshot, nil
}

func snapshotInput(name string) bool {
	if extension := path.Ext(name); extension == ".proto" || extension == ".yaml" || extension == ".yml" {
		return true
	}
	switch path.Base(name) {
	case "protobuf.mod", "protobuf.lock", "easyp.yaml", "buf.yaml", "buf.yml", "buf.work.yaml":
		return true
	default:
		return false
	}
}

func snapshotAdditionalPolicy(name string) bool {
	if path.Ext(name) != ".yaml" && path.Ext(name) != ".yml" {
		return false
	}
	switch path.Base(name) {
	case "easyp.yaml", "buf.yaml", "buf.yml", "buf.work.yaml":
		return false
	}
	return true
}
