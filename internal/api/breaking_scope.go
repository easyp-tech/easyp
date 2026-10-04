package api

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	gitadapter "github.com/easyp-tech/easyp/internal/adapters/go_git"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	disk "github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/modules"
)

type breakingScope struct{ files []string }

// discoverBreakingScopes assigns files to the nearest module in this revision.
// Repository-relative paths make findings independent of --root and import roots.
// root is the scanned tree; repositoryRoot identifies absolute replacement paths.
func discoverBreakingScopes(root, repositoryRoot, scanRelative string) (map[string]breakingScope, error) {
	scopes := make(map[string]breakingScope)
	scan := filepath.Join(root, scanRelative)
	if _, err := os.Stat(scan); os.IsNotExist(err) {
		return scopes, nil
	} else if err != nil {
		return nil, fmt.Errorf("Stat: %w", err)
	}
	manifests := make(map[string]v1.Module)
	err := filepath.WalkDir(scan, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if policySourceExcluded(root, path) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || filepath.Ext(path) != ".proto" {
			return nil
		}
		unselectedReplacement, err := isUnselectedReplacementSource(root, repositoryRoot, scan, path)
		if err != nil {
			return fmt.Errorf("isUnselectedReplacementSource: %w", err)
		}
		if unselectedReplacement {
			return nil
		}
		directory, err := findV1PolicyModuleDir(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		key := "."
		if directory != "" {
			module, known := manifests[directory]
			if !known {
				_, module, err = modules.ReadManifest(directory)
				if err != nil {
					return err
				}
				manifests[directory] = module
			}
			owned := false
			for _, sourceRoot := range module.Roots {
				if path_helpers.IsTargetPath(filepath.Join(directory, sourceRoot), path) {
					owned = true
					break
				}
			}
			if !owned {
				return nil
			}
			key, err = filepath.Rel(root, directory)
			if err != nil {
				return err
			}
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		scope := scopes[key]
		scope.files = append(scope.files, filepath.ToSlash(relative))
		scopes[key] = scope
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("WalkDir: %w", err)
	}
	return scopes, nil
}

// scopedBreakingWalker exposes target files by repository path, but does not
// let import lookup accidentally read another module from the repository root.
type scopedBreakingWalker struct {
	core.FS
	root    string
	files   []string
	allowed map[string]bool
}

func newBreakingWalker(root string, files []string) *scopedBreakingWalker {
	allowed := make(map[string]bool, len(files))
	for _, file := range files {
		allowed[filepath.ToSlash(file)] = true
	}
	return &scopedBreakingWalker{FS: disk.NewFSWalker(root, "."), root: root, files: files, allowed: allowed}
}

func (w *scopedBreakingWalker) RootPath() string { return w.root }
func (w *scopedBreakingWalker) Open(name string) (io.ReadCloser, error) {
	if !w.allowed[filepath.ToSlash(name)] {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}
	return w.FS.Open(name)
}
func (w *scopedBreakingWalker) WalkDir(visit func(string, error) error) error {
	for _, file := range w.files {
		if err := visit(file, nil); err != nil {
			return err
		}
	}
	return nil
}

func breakingImportRoots(ctx context.Context, cache modules.Cache, root, moduleRelative string, files []string, snapshot *gitadapter.Snapshot) ([]string, error) {
	return breakingImportRootsMode(ctx, cache, root, moduleRelative, files, snapshot, false)
}

func breakingImportRootsMode(ctx context.Context, cache modules.Cache, root, moduleRelative string, files []string, snapshot *gitadapter.Snapshot, frozen bool) ([]string, error) {
	directory := filepath.Join(root, moduleRelative)
	if len(files) == 0 {
		// A deleted module has no graph in this revision, but an explicit empty
		// module still requires its own lock and replacement validation.
		if !frozen {
			return nil, nil
		}
		if _, err := os.Stat(filepath.Join(directory, v1.ModuleFile)); os.IsNotExist(err) {
			return nil, nil
		}
	}
	if frozen {
		return policyImportRoots(ctx, cache, directory, true)
	}
	if _, err := os.Stat(filepath.Join(directory, v1.ModuleFile)); os.IsNotExist(err) {
		return []string{directory}, nil
	} else if err != nil {
		return nil, fmt.Errorf("Stat: %w", err)
	}
	if snapshot == nil {
		return ensureV1PolicyImportRoots(ctx, cache, directory)
	}
	_, module, err := modules.ReadManifest(directory)
	if err != nil {
		return nil, fmt.Errorf("ReadManifest: %w", err)
	}
	own, err := modules.ModuleSources(directory, module)
	if err != nil {
		return nil, fmt.Errorf("ModuleSources: %w", err)
	}
	var dependencies modules.SourceRoots
	if len(module.Replaces) > 0 {
		graph, err := modules.EnsureEffectiveGraph(ctx, directory, module, cache, func(target string) (string, error) {
			return snapshotReplacementPath(snapshot, directory, target)
		})
		if err != nil {
			return nil, fmt.Errorf("EnsureEffectiveGraph: %w", err)
		}
		dependencies = graph.Sources
	} else {
		dependencies, err = modules.EnsureSources(ctx, directory, module, cache)
		if err != nil {
			return nil, fmt.Errorf("EnsureSources: %w", err)
		}
	}

	if err := modules.CheckImportCollisions(directory, module.Roots, dependencies.Paths()); err != nil {
		return nil, fmt.Errorf("CheckImportCollisions: %w", err)
	}
	return append(own.Paths(), dependencies.Paths()...), nil
}

// Local replacements inside the repository are read from the baseline snapshot.
// External working directories have no historical identity and cannot safely
// stand in for a dependency at the baseline revision.
func snapshotReplacementPath(snapshot *gitadapter.Snapshot, directory, target string) (string, error) {
	resolved := modules.ResolveReplacementPath(directory, target)
	if filepath.IsAbs(target) {
		relative, err := baselineRepositoryRelative(snapshot.RepositoryRoot, target)
		if err != nil || !filepath.IsLocal(relative) {
			return "", fmt.Errorf("baseline replacement points outside the Git repository: %q; use a locked dependency for a reproducible baseline", target)
		}
		resolved = filepath.Join(snapshot.Root, relative)
	}
	relative, err := filepath.Rel(snapshot.Root, resolved)
	if err != nil || !filepath.IsLocal(relative) {
		return "", fmt.Errorf("baseline replacement leaves the Git snapshot: %q", target)
	}
	// Do not follow a snapshot symlink into present-day working files.
	canonical, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		return "", fmt.Errorf("EvalSymlinks: %w", err)
	}
	snapshotRoot, err := filepath.EvalSymlinks(snapshot.Root)
	if err != nil {
		return "", fmt.Errorf("EvalSymlinks: %w", err)
	}
	relative, err = filepath.Rel(snapshotRoot, canonical)
	if err != nil || !filepath.IsLocal(relative) {
		return "", fmt.Errorf("baseline replacement leaves the Git snapshot: %q", target)
	}
	return resolved, nil
}

// baselineRepositoryRelative accepts filesystem aliases such as macOS /tmp,
// including paths deleted from the current tree but present at the baseline.
func baselineRepositoryRelative(root, target string) (string, error) {
	relative, err := filepath.Rel(root, target)
	if err == nil && filepath.IsLocal(relative) {
		return relative, nil
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("EvalSymlinks: %w", err)
	}
	ancestor := target
	var suffix []string
	for {
		resolved, err := filepath.EvalSymlinks(ancestor)
		if err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return filepath.Rel(realRoot, resolved)
		}
		if !os.IsNotExist(err) || filepath.Dir(ancestor) == ancestor {
			return "", fmt.Errorf("EvalSymlinks: %w", err)
		}
		suffix = append(suffix, filepath.Base(ancestor))
		ancestor = filepath.Dir(ancestor)
	}
}
