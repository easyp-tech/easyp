package go_git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"

	"github.com/easyp-tech/easyp/internal/adapters/gitsnapshot"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/sourceview"
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
	tree, err := gitsnapshot.New(repository, commit.Hash)
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
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
	view := sourceview.New(tree)
	err = view.Walk(ctx, ".", func(name string, resolved sourceview.Resolution, walkErr error) error {
		if walkErr != nil {
			if !snapshotInput(name) || snapshotAdditionalPolicy(name) {
				return nil
			}
			return fmt.Errorf("baseline input alias %q: %w", name, walkErr)
		}
		if resolved.Info.IsDir() {
			return nil
		}
		if !snapshotInput(name) {
			return nil
		}
		relative := filepath.FromSlash(name)
		if !filepath.IsLocal(relative) {
			return fmt.Errorf("invalid snapshot path %q", name)
		}
		reader, err := view.Open(ctx, name)
		if err != nil {
			return fmt.Errorf("Open: %w", err)
		}
		raw, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return fmt.Errorf("ReadAll: %w", readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("Close: %w", closeErr)
		}
		if err := gitsnapshot.ValidateDestination(root, name); err != nil {
			return fmt.Errorf("ValidateDestination: %w", err)
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
		return nil, fmt.Errorf("Walk: %w", err)
	}
	if err := validateSnapshotRoots(ctx, view, root); err != nil {
		return nil, fmt.Errorf("validateSnapshotRoots: %w", err)
	}
	return snapshot, nil
}

// Required roots come from revision-local manifests, not filename extensions.
// A failed extensionless root alias must not silently create an empty baseline.
func validateSnapshotRoots(ctx context.Context, view *sourceview.View, root string) error {
	err := filepath.WalkDir(root, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != v1.ModuleFile {
			return nil
		}
		raw, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("ReadFile: %w", err)
		}
		if !v1.IsModuleManifest(raw) {
			return nil
		}
		module, err := v1.ParseModule(strings.NewReader(string(raw)))
		if err != nil {
			return fmt.Errorf("ParseModule: %w", err)
		}
		relative, err := filepath.Rel(root, filepath.Dir(filename))
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		for _, declared := range module.Roots {
			logical := path.Join(filepath.ToSlash(relative), filepath.ToSlash(declared))
			resolved, err := view.Resolve(ctx, logical)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) && len(resolved.Links) == 0 {
					continue
				}
				return fmt.Errorf("Resolve: baseline root %q: %w", logical, err)
			}
			if !resolved.Info.IsDir() {
				return fmt.Errorf("baseline root %q is not a directory", logical)
			}
			if err := view.Walk(ctx, logical, func(name string, resolved sourceview.Resolution, walkErr error) error {
				selectedRoot := filepath.Join(root, filepath.FromSlash(logical))
				selectedPath := filepath.Join(root, filepath.FromSlash(name))
				if path_helpers.HiddenOrVendorSourcePath(selectedRoot, selectedPath) || path_helpers.ShouldSkipV1SourceDir(selectedRoot, selectedPath) {
					if resolved.Info != nil && resolved.Info.IsDir() {
						return fs.SkipDir
					}
					return nil
				}
				if errors.Is(walkErr, gitsnapshot.ErrGitlink) {
					if name != logical && len(resolved.Links) == 0 {
						return nil
					}
					return walkErr
				}
				if walkErr != nil && errors.Is(walkErr, sourceview.ErrCycle) && resolved.Info != nil && resolved.Info.IsDir() {
					return fmt.Errorf("baseline directory alias %q: %w", name, walkErr)
				}
				return nil
			}); err != nil {
				return fmt.Errorf("Walk: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("WalkDir: %w", err)
	}
	return nil
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
