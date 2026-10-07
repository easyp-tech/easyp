package gitmodules

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

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"

	"github.com/easyp-tech/easyp/internal/adapters/gitsnapshot"
	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

// prepareV1Snapshot constructs the regular logical view of one pinned Git tree.
// Metadata is materialized before passing it to the existing metadata readers.
// The caller owns and removes the returned directory.
func prepareV1Snapshot(ctx context.Context, checkout, commit, source, subdir string, roots []string, inspect bool) (directory string, module v1.Module, inspection *modules.RootInspection, err error) {
	tree, err := sourceV1SnapshotTree(ctx, checkout, commit)
	if err != nil {
		return "", v1.Module{}, nil, fmt.Errorf("sourceV1SnapshotTree: %w", err)
	}
	view := sourceview.New(tree)
	directory, err = os.MkdirTemp(filepath.Dir(checkout), "snapshot-*")
	if err != nil {
		return "", v1.Module{}, nil, fmt.Errorf("MkdirTemp: %w", err)
	}
	stagePath := directory
	defer func() {
		if err != nil {
			removeErr := os.RemoveAll(stagePath)
			if removeErr != nil {
				err = errors.Join(err, fmt.Errorf("RemoveAll: %w", removeErr))
			}
		}
	}()
	inventory, err := stageSnapshotMetadata(ctx, tree, view, directory, subdir)
	if err != nil {
		return "", v1.Module{}, nil, fmt.Errorf("stageSnapshotMetadata: %w", err)
	}
	aliases, regular, metadataFailures := inventory.aliases, inventory.regular, inventory.metadataFailures
	module, err = moduleconfig.ReadGitDependencyAt(directory, source, subdir)
	if err != nil {
		for _, failure := range metadataFailures {
			if strings.Contains(err.Error(), filepath.Join(directory, filepath.FromSlash(failure.name))) {
				err = errors.Join(err, fmt.Errorf("Resolve: %s: %w", failure.name, failure.err))
			}
		}
		return "", v1.Module{}, nil, fmt.Errorf("ReadGitDependencyAt: %w", err)
	}
	module, err = applyV1ModuleRoots(module, roots)
	if err != nil {
		return "", v1.Module{}, nil, fmt.Errorf("applyV1ModuleRoots: %w", err)
	}
	provisional := inspect && !module.RootsFromMetadata
	ignoredMetadata := make(map[string]bool, len(metadataFailures))
	for _, failure := range metadataFailures {
		if (!provisional && snapshotFailedMetadataOwnsSources(failure.name, directory, module)) || (failure.unstaged && snapshotNativeCandidate(failure.name, source, subdir)) {
			return "", v1.Module{}, nil, fmt.Errorf("Resolve: %s: %w", failure.name, failure.err)
		}
		ignoredMetadata[failure.name] = true
		if failure.unstaged {
			continue
		}
		if err := os.Remove(filepath.Join(directory, filepath.FromSlash(failure.name))); err != nil {
			return "", v1.Module{}, nil, fmt.Errorf("Remove: %w", err)
		}
		if err := pruneEmptySnapshotParents(directory, failure.name); err != nil {
			return "", v1.Module{}, nil, fmt.Errorf("pruneEmptySnapshotParents: %w", err)
		}
	}
	// Stage ordinary files only after metadata has selected Buf source paths.
	// Discarded proto names must not collide with retained names on the host.
	for _, name := range regular {
		if path.Ext(name) == ".proto" && len(selectV1ProtoFiles([]string{name}, module.ProtoFilters)) == 0 {
			continue
		}
		if err := materializeSnapshotFile(ctx, view, directory, name); err != nil {
			return "", v1.Module{}, nil, fmt.Errorf("materializeSnapshotFile: %w", err)
		}
	}
	if err := materializeSnapshotAuxiliaryAliases(ctx, view, directory, aliases); err != nil {
		return "", v1.Module{}, nil, fmt.Errorf("materializeSnapshotAuxiliaryAliases: %w", err)
	}
	if !provisional {
		if err := materializeSelectedSnapshotAliases(ctx, view, directory, module, ignoredMetadata); err != nil {
			return "", v1.Module{}, nil, fmt.Errorf("materializeSelectedSnapshotAliases: %w", err)
		}
	}
	if inspect {
		inspection, err = inspectV1Snapshot(ctx, view, directory, module, commit)
		if err != nil {
			return "", v1.Module{}, nil, fmt.Errorf("inspectV1Snapshot: %w", err)
		}
		inspection.Provisional = provisional
	}
	// Buf filters select the same logical paths for hashing and installation.
	err = filepath.WalkDir(directory, func(file string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(file) != ".proto" {
			return nil
		}
		name, err := filepath.Rel(directory, file)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		if len(selectV1ProtoFiles([]string{filepath.ToSlash(name)}, module.ProtoFilters)) == 0 {
			return os.Remove(file)
		}
		return nil
	})
	if err != nil {
		return "", v1.Module{}, nil, fmt.Errorf("WalkDir: %w", err)
	}
	return directory, module, inspection, nil
}

type snapshotInventory struct {
	aliases          []string
	regular          []string
	metadataFailures []snapshotMetadataFailure
}

// pruneEmptySnapshotParents removes namespaces created only by ignored markers.
// Retained metadata stops pruning; later source files choose their own spelling.
func pruneEmptySnapshotParents(directory, name string) error {
	for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
		filename := filepath.Join(directory, filepath.FromSlash(parent))
		entries, err := os.ReadDir(filename)
		if err != nil {
			return fmt.Errorf("ReadDir: %w", err)
		}
		if len(entries) > 0 {
			return nil
		}
		if err := os.Remove(filename); err != nil {
			return fmt.Errorf("Remove: %w", err)
		}
	}
	return nil
}

// stageSnapshotMetadata discovers immutable metadata before source filtering.
// Failed markers preserve ownership and parser precedence without pointer reads.
func stageSnapshotMetadata(ctx context.Context, tree *gitsnapshot.FS, view *sourceview.View, directory, subdir string) (snapshotInventory, error) {
	var aliases, regular []string
	var metadataFailures []snapshotMetadataFailure
	failedMarker := func(name string, cause error) error {
		failure := snapshotMetadataFailure{name: name, err: cause}
		err := makeSnapshotDirectory(directory, name)
		if errors.Is(err, gitsnapshot.ErrPathCollision) {
			// A failed alias has no retained bytes. Keep its witness separate
			// if the host cannot represent it beside staged metadata.
			failure.unstaged = true
		} else if err != nil {
			return fmt.Errorf("makeSnapshotDirectory: %w", err)
		}
		metadataFailures = append(metadataFailures, failure)
		return nil
	}
	err := fs.WalkDir(tree, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, gitsnapshot.ErrGitlink) {
				return fs.SkipDir
			}
			return walkErr
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			aliases = append(aliases, name)
			return nil
		}
		if entry.IsDir() {
			_, statErr := tree.Lstat(name)
			if statErr != nil && errors.Is(statErr, gitsnapshot.ErrGitlink) {
				return fs.SkipDir
			}
			if statErr != nil {
				return statErr
			}
			return nil
		}
		if !moduleconfig.IsGitDependencyConfigFile(path.Base(name)) {
			regular = append(regular, name)
			return nil
		}
		return materializeSnapshotFile(ctx, view, directory, name)
	})
	if err != nil {
		return snapshotInventory{}, fmt.Errorf("WalkDir: %w", err)
	}
	// Traverse directory aliases for metadata discovery. Irrelevant dangling or
	// recursive aliases are omitted; roots selected by metadata are strict below.
	for _, name := range aliases {
		if snapshotMetadataExcluded(name, subdir) {
			continue
		}
		resolution, resolveErr := view.Resolve(ctx, name)
		if resolveErr != nil {
			if moduleconfig.IsGitDependencyConfigFile(path.Base(name)) {
				// A directory records presence without feeding a pointer or target bytes
				// to a reader. Existing metadata precedence decides whether it is read.
				if err := failedMarker(name, resolveErr); err != nil {
					return snapshotInventory{}, err
				}
			}
			continue
		}
		if resolution.Info.IsDir() && moduleconfig.IsGitDependencyConfigFile(path.Base(name)) {
			if err := failedMarker(name, sourceview.ErrUnsupported); err != nil {
				return snapshotInventory{}, err
			}
			continue
		}
		if resolution.Info.IsDir() {
			walkErr := view.Walk(ctx, name, func(logical string, resolved sourceview.Resolution, walkErr error) error {
				if walkErr != nil {
					if moduleconfig.IsGitDependencyConfigFile(path.Base(logical)) && !snapshotMetadataExcluded(logical, subdir) {
						if err := failedMarker(logical, walkErr); err != nil {
							return err
						}
					}
					return nil
				}
				if resolved.Info.IsDir() {
					if snapshotMetadataExcluded(logical, subdir) {
						return fs.SkipDir
					}
					return nil
				}
				if moduleconfig.IsGitDependencyConfigFile(path.Base(logical)) && !snapshotMetadataExcluded(logical, subdir) {
					return materializeSnapshotFile(ctx, view, directory, logical)
				}
				return nil
			})
			if walkErr != nil {
				return snapshotInventory{}, fmt.Errorf("Walk: %w", walkErr)
			}
			continue
		}
		if moduleconfig.IsGitDependencyConfigFile(path.Base(name)) {
			if err := materializeSnapshotFile(ctx, view, directory, name); err != nil {
				return snapshotInventory{}, fmt.Errorf("materializeSnapshotFile: %w", err)
			}
		}
	}
	return snapshotInventory{aliases: aliases, regular: regular, metadataFailures: metadataFailures}, nil
}

// materializeSelectedSnapshotAliases validates declared roots and copies aliases
// selected by source ownership and Buf filters. Auxiliary links are handled separately.
func materializeSelectedSnapshotAliases(ctx context.Context, view *sourceview.View, directory string, module v1.Module, ignoredMetadata map[string]bool) error {
	for _, root := range module.Roots {
		logicalRoot := filepath.ToSlash(root)
		resolution, resolveErr := view.Resolve(ctx, logicalRoot)
		if resolveErr != nil {
			if errors.Is(resolveErr, fs.ErrNotExist) && len(resolution.Links) == 0 {
				if err := makeSnapshotDirectory(directory, root); err != nil {
					return fmt.Errorf("MkdirAll: %w", err)
				}
				continue
			}
			return fmt.Errorf("Resolve: root %q: %w", root, resolveErr)
		}
		if !resolution.Info.IsDir() {
			return fmt.Errorf("module %s has invalid root %q: not a directory", module.Name, root)
		}
		if err := makeSnapshotDirectory(directory, root); err != nil {
			return fmt.Errorf("MkdirAll: %w", err)
		}
		err := view.Walk(ctx, logicalRoot, func(logical string, resolved sourceview.Resolution, walkErr error) error {
			if ignoredMetadata[logical] {
				return nil
			}
			if walkErr != nil {
				if logical != logicalRoot && snapshotExcludedDirectory(logical, logicalRoot, directory, module) {
					return nil
				}
				// An unentered submodule is an opaque boundary, like a nested
				// module. Explicit roots and aliases cannot cross that boundary.
				if errors.Is(walkErr, gitsnapshot.ErrGitlink) && logical != logicalRoot && len(resolved.Links) == 0 {
					return nil
				}
				if snapshotSelectedAlias(logical, module) || snapshotStrictDirectory(logical, module) || (resolved.Info != nil && resolved.Info.IsDir()) || errors.Is(walkErr, gitsnapshot.ErrGitlink) {
					return walkErr
				}
				return nil
			}
			if resolved.Info.IsDir() {
				if logical != logicalRoot && snapshotExcludedDirectory(logical, logicalRoot, directory, module) {
					return fs.SkipDir
				}
				return nil
			}
			if len(resolved.Links) == 0 {
				return nil
			}
			if !snapshotSelectedAlias(logical, module) {
				return nil
			}
			return materializeSnapshotFile(ctx, view, directory, logical)
		})
		if err != nil {
			return fmt.Errorf("Walk: root %q: %w", root, err)
		}
	}
	return nil
}

func sourceV1SnapshotView(ctx context.Context, checkout, commit string) (*sourceview.View, error) {
	tree, err := sourceV1SnapshotTree(ctx, checkout, commit)
	if err != nil {
		return nil, fmt.Errorf("sourceV1SnapshotTree: %w", err)
	}
	return sourceview.New(tree), nil
}

func sourceV1SnapshotTree(ctx context.Context, checkout, commit string) (*gitsnapshot.FS, error) {
	if len(commit) == 64 {
		tree, err := gitsnapshot.NewRepository(ctx, checkout, commit)
		if err != nil {
			return nil, fmt.Errorf("NewRepository: %w", err)
		}
		return tree, nil
	}
	storage := filesystem.NewStorageWithOptions(osfs.New(filepath.Join(checkout, ".git")), cache.NewObjectLRUDefault(), filesystem.Options{AlternatesFS: osfs.New(string(filepath.Separator))})
	repo, err := git.Open(storage, osfs.New(checkout))
	if err != nil {
		return nil, fmt.Errorf("Open: %w", err)
	}
	tree, err := gitsnapshot.New(repo, plumbing.NewHash(commit))
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}
	return tree, nil
}

type snapshotMetadataFailure struct {
	name     string
	err      error
	unstaged bool
}

// Native candidate manifests are read even when another candidate matched.
// Buf/EasyP fallback metadata has no role once a native manifest wins.
func snapshotNativeCandidate(name, source, subdir string) bool {
	if path.Base(name) != v1.ModuleFile {
		return false
	}
	directory := path.Dir(name)
	if directory == path.Clean(subdir) || (subdir == "" && directory == ".") {
		return true
	}
	major, err := v1.ModulePathMajor(source)
	return err == nil && strings.HasPrefix(major, "/") && directory == path.Join(subdir, strings.TrimPrefix(major, "/"))
}

func snapshotMetadataExcluded(name, subdir string) bool {
	directory := path.Dir(name)
	base := filepath.ToSlash(subdir)
	if base != "" && base != "." && v1FileWithin(directory, base) {
		directory = strings.TrimPrefix(strings.TrimPrefix(directory, base), "/")
	}
	for _, component := range strings.Split(directory, "/") {
		if component != "." && (strings.HasPrefix(component, ".") || component == "easyp_vendor") {
			return true
		}
	}
	return false
}

func snapshotFailedMetadataOwnsSources(name, directory string, module v1.Module) bool {
	switch path.Base(name) {
	case v1.ModuleFile, "buf.yaml", "buf.work.yaml":
	default:
		return false
	}
	parent := path.Dir(name)
	for _, root := range module.Roots {
		root = filepath.ToSlash(root)
		if parent == root || !v1FileWithin(parent, root) {
			continue
		}
		excluded := false
		for ancestor := parent; ancestor != root && v1FileWithin(ancestor, root); ancestor = path.Dir(ancestor) {
			if snapshotExcludedDirectory(ancestor, root, directory, module) {
				excluded = true
				break
			}
		}
		if !excluded {
			return true
		}
	}
	return false
}

// materializeSnapshotAuxiliaryAliases retains bounded regular non-proto aliases.
// Consumers can select arbitrary policy fragments; Fetch has no consumer context
// with which to narrow those paths. Failed auxiliary aliases remain omitted.
func materializeSnapshotAuxiliaryAliases(ctx context.Context, view *sourceview.View, directory string, aliases []string) error {
	for _, name := range aliases {
		resolved, err := view.Resolve(ctx, name)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		if resolved.Info.IsDir() {
			err = view.Walk(ctx, name, func(logical string, resolved sourceview.Resolution, walkErr error) error {
				if walkErr != nil {
					return nil
				}
				if resolved.Info.IsDir() || path.Ext(logical) == ".proto" {
					return nil
				}
				return materializeSnapshotFile(ctx, view, directory, logical)
			})
			if err != nil {
				return fmt.Errorf("Walk: %w", err)
			}
			continue
		}
		if path.Ext(name) == ".proto" {
			continue
		}
		if err := materializeSnapshotFile(ctx, view, directory, name); err != nil {
			return fmt.Errorf("materializeSnapshotFile: %w", err)
		}
	}
	return nil
}

func snapshotSelectedAlias(name string, module v1.Module) bool {
	if moduleconfig.IsGitDependencyConfigFile(path.Base(name)) {
		return true
	}
	if path.Ext(name) != ".proto" {
		return false
	}
	for _, root := range module.Roots {
		if v1FileWithin(name, root) {
			return len(selectV1ProtoFiles([]string{name}, module.ProtoFilters)) != 0
		}
	}
	return false
}

func snapshotStrictDirectory(name string, module v1.Module) bool {
	for _, root := range module.Roots {
		if name == filepath.ToSlash(root) {
			return true
		}
	}
	for _, filter := range module.ProtoFilters {
		for _, included := range filter.Includes {
			if name == filepath.ToSlash(included) {
				return true
			}
		}
	}
	return false
}

func snapshotExcludedDirectory(name, root, directory string, module v1.Module) bool {
	relative := name
	if name == root {
		relative = ""
	} else if root != "." {
		relative = strings.TrimPrefix(name, root+"/")
	}
	for _, component := range strings.Split(relative, "/") {
		if strings.HasPrefix(component, ".") || component == "easyp_vendor" {
			return true
		}
	}
	// Nested native modules retain their own ownership.
	if name != root {
		data, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(name), v1.ModuleFile))
		if err == nil && v1.IsModuleManifest(data) {
			return true
		}
	}
	if len(module.ProtoFilters) == 0 {
		return false
	}
	for _, filter := range module.ProtoFilters {
		if !v1FileWithin(name, filter.Root) {
			continue
		}
		allowed := len(filter.Includes) == 0
		for _, included := range filter.Includes {
			allowed = allowed || v1FileWithin(name, included) || v1FileWithin(included, name)
		}
		for _, excluded := range filter.Excludes {
			if v1FileWithin(name, excluded) {
				allowed = false
				break
			}
		}
		if allowed {
			return false
		}
	}
	return true
}

func materializeSnapshotFile(ctx context.Context, view *sourceview.View, directory, name string) error {
	file, err := view.Open(ctx, name)
	if err != nil {
		return fmt.Errorf("Open: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("Stat: %w", err)
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return fmt.Errorf("non-regular snapshot file %q", name)
	}
	destination := filepath.Join(directory, filepath.FromSlash(name))
	if err := gitsnapshot.ValidateDestination(directory, name); err != nil {
		closeErr := file.Close()
		return errors.Join(fmt.Errorf("ValidateDestination: %w", err), closeErr)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		_ = file.Close()
		return fmt.Errorf("MkdirAll: %w", err)
	}
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("OpenFile: %w", err)
	}
	_, copyErr := io.Copy(output, file)
	closeErr := errors.Join(file.Close(), output.Close())
	if err := errors.Join(copyErr, closeErr, ctx.Err()); err != nil {
		return fmt.Errorf("Copy: %w", err)
	}
	return nil
}

func makeSnapshotDirectory(directory, name string) error {
	if err := gitsnapshot.ValidateDestination(directory, name); err != nil {
		return fmt.Errorf("ValidateDestination: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(directory, filepath.FromSlash(name)), 0o755); err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}
	return nil
}

func snapshotV1Files(directory string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(directory, func(file string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("Info: %w", err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("non-regular snapshot file %q", file)
		}
		name, err := filepath.Rel(directory, file)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		files = append(files, filepath.ToSlash(name))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("WalkDir: %w", err)
	}
	return files, nil
}

func hashSnapshotV1Files(directory string) (string, error) {
	files, err := snapshotV1Files(directory)
	if err != nil {
		return "", fmt.Errorf("snapshotV1Files: %w", err)
	}
	hash, err := hashV1Files(directory, files)
	if err != nil {
		return "", fmt.Errorf("hashV1Files: %w", err)
	}
	return hash, nil
}
