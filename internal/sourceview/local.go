package sourceview

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

// ErrNestedRepository marks an alias crossing into a nested local repository.
var ErrNestedRepository = errors.New("source alias crosses nested repository")

// OpenLocal opens a logical source through a bounded local root. Closing the
// returned file also closes the root handle.
func OpenLocal(ctx context.Context, rootPath, logical string) (fs.File, error) {
	return OpenLocalSelected(ctx, rootPath, logical, nil)
}

// OpenLocalSelected validates a resolved target before opening its physical
// regular path and checks that the opened inode is the selected target.
// An optional nestedAllowed predicate must prove explicit source ownership for
// a target crossing a nested repository boundary.
func OpenLocalSelected(ctx context.Context, rootPath, logical string, selected func(Resolution) bool, nestedAllowed ...func(Resolution) bool) (fs.File, error) {
	root, view, err := localView(rootPath)
	if err != nil {
		return nil, fmt.Errorf("localView: %w", err)
	}
	openFile := func(name string) (fs.File, error) {
		return root.Open(filepath.FromSlash(name))
	}
	lstat := func(name string) (fs.FileInfo, error) {
		return root.Lstat(filepath.FromSlash(name))
	}
	file, err := openSelected(ctx, openFile, lstat, view, logical, selected, nestedAllowed...)
	if err != nil {
		return nil, errors.Join(err, root.Close())
	}
	return &localFile{File: file, root: root}, nil
}

// openSelected resolves and opens logical below a bounded root owned by the caller.
func openSelected(
	ctx context.Context,
	openFile func(string) (fs.File, error),
	lstat func(string) (fs.FileInfo, error),
	view *View,
	logical string,
	selected func(Resolution) bool,
	nestedAllowed ...func(Resolution) bool,
) (fs.File, error) {
	resolved, err := view.Resolve(ctx, filepath.ToSlash(logical))
	if err != nil {
		return nil, fmt.Errorf("Resolve: %w", localResolutionError(logical, resolved, err))
	}
	if err := checkLocalResolution(lstat, resolved); err != nil {
		permitted := errors.Is(err, ErrNestedRepository) && len(nestedAllowed) > 0 && nestedAllowed[0] != nil && nestedAllowed[0](resolved)
		if !permitted {
			return nil, err
		}
	}
	if !resolved.Info.Mode().IsRegular() {
		return nil, ErrUnsupported
	}
	if selected != nil && !selected(resolved) {
		return nil, &fs.PathError{Op: "open", Path: logical, Err: fs.ErrNotExist}
	}
	file, err := openFile(resolved.Path)
	if err != nil {
		return nil, fmt.Errorf("Open: %w", err)
	}
	opened, err := file.Stat()
	if err != nil {
		return nil, errors.Join(fmt.Errorf("Stat: %w", err), file.Close())
	}
	if !os.SameFile(resolved.Info, opened) || resolved.Info.Mode() != opened.Mode() || resolved.Info.Size() != opened.Size() || !resolved.Info.ModTime().Equal(opened.ModTime()) {
		return nil, errors.Join(ErrChanged, file.Close())
	}
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(err, file.Close())
	}
	return file, nil
}

type localFile struct {
	fs.File
	root *os.Root
}

func (file *localFile) Close() error { return errors.Join(file.File.Close(), file.root.Close()) }

// ReadLocal reads a logical regular source without leaving its local root.
func ReadLocal(ctx context.Context, rootPath, logical string) (_ []byte, resultErr error) {
	file, err := OpenLocal(ctx, rootPath, logical)
	if err != nil {
		return nil, fmt.Errorf("OpenLocal: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, file.Close()) }()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("Stat: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("source %q is not a regular file", logical)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %w", err)
	}
	return data, nil
}

// WalkLocal walks a bounded local tree while preserving logical source names.
func WalkLocal(ctx context.Context, rootPath, logical string, visit WalkFunc) (resultErr error) {
	root, view, err := localView(rootPath)
	if err != nil {
		return fmt.Errorf("localView: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, root.Close()) }()
	return view.Walk(ctx, filepath.ToSlash(logical), func(name string, resolved Resolution, err error) error {
		if err == nil {
			err = CheckLocalResolution(root, resolved)
		}
		return visit(name, resolved, err)
	})
}

// ResolveLocal resolves a logical source without leaving its local root.
func ResolveLocal(ctx context.Context, rootPath, logical string) (_ Resolution, resultErr error) {
	root, view, err := localView(rootPath)
	if err != nil {
		return Resolution{}, fmt.Errorf("localView: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, root.Close()) }()
	resolved, err := view.Resolve(ctx, filepath.ToSlash(logical))
	if err == nil {
		err = CheckLocalResolution(root, resolved)
	}
	return resolved, localResolutionError(logical, resolved, err)
}

// ResolveLocalOptional supports absent import candidates beneath directory
// aliases while preserving the local boundary and nested-repository checks.
func ResolveLocalOptional(ctx context.Context, rootPath, logical string) (_ Resolution, resultErr error) {
	root, view, err := localView(rootPath)
	if err != nil {
		return Resolution{}, fmt.Errorf("localView: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, root.Close()) }()
	resolved, err := view.ResolveOptional(ctx, filepath.ToSlash(logical))
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		checkErr := CheckLocalResolution(root, resolved)
		if checkErr != nil {
			return resolved, checkErr
		}
	}
	return resolved, err
}

func openLocalRoot(rootPath string) (*os.Root, string, string, error) {
	canonical, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return nil, "", "", fmt.Errorf("EvalSymlinks: %w", err)
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return nil, "", "", fmt.Errorf("Abs: %w", err)
	}
	root, err := os.OpenRoot(canonical)
	if err != nil {
		return nil, "", "", fmt.Errorf("OpenRoot: %w", err)
	}
	requested, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, "", "", errors.Join(fmt.Errorf("Abs: %w", err), root.Close())
	}
	original, err := os.Stat(requested)
	if err != nil {
		return nil, "", "", errors.Join(fmt.Errorf("Stat: %w", err), root.Close())
	}
	opened, err := root.Stat(".")
	if err != nil {
		return nil, "", "", errors.Join(fmt.Errorf("Stat: %w", err), root.Close())
	}
	if !os.SameFile(original, opened) {
		return nil, "", "", errors.Join(ErrChanged, root.Close())
	}
	return root, canonical, requested, nil
}

func localView(rootPath string) (*os.Root, *View, error) {
	root, canonical, requested, err := openLocalRoot(rootPath)
	if err != nil {
		return nil, nil, err
	}
	view, err := NewLocal(root.FS(), canonical, requested)
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("NewLocal: %w", err), root.Close())
	}
	return root, view, nil
}

func localTreeView(rootPath string, cache *dirCache) (*rootTree, *View, error) {
	root, canonical, requested, err := openLocalRoot(rootPath)
	if err != nil {
		return nil, nil, err
	}
	tree := newRootTree(root, cache)
	view, err := NewLocal(tree, canonical, requested)
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("NewLocal: %w", err), tree.Close())
	}
	return tree, view, nil
}

func localResolutionError(logical string, resolved Resolution, err error) error {
	if errors.Is(err, fs.ErrNotExist) && len(resolved.Links) > 0 {
		return fmt.Errorf("source alias %q is dangling: %v: %w", logical, err, ErrUnsupported)
	}
	return err
}

// CheckLocalResolution rejects alias hops and targets inside nested repositories.
// The boundary root's own Git marker is allowed. Graph readers may authorize a
// target through a separately declared source boundary using OpenLocalSelected.
func CheckLocalResolution(root *os.Root, resolved Resolution) error {
	return checkLocalResolution(func(name string) (fs.FileInfo, error) {
		return root.Lstat(filepath.FromSlash(name))
	}, resolved)
}

// checkLocalResolution takes slash-separated names relative to the root.
func checkLocalResolution(lstat func(string) (fs.FileInfo, error), resolved Resolution) error {
	if len(resolved.Links) == 0 {
		return nil
	}
	directories := []string{path.Dir(resolved.Path)}
	if resolved.Info != nil && resolved.Info.IsDir() {
		directories[0] = resolved.Path
	}
	for _, link := range resolved.Links {
		directories = append(directories, path.Dir(link.Path))
	}
	seen := make(map[string]bool)
	for _, directory := range directories {
		for current := directory; current != "."; current = path.Dir(current) {
			if seen[current] {
				break
			}
			seen[current] = true
			_, err := lstat(path.Join(current, ".git"))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return fmt.Errorf("Lstat: %w", err)
			}
			return fmt.Errorf("source target %q crosses nested repository %q: %w", resolved.Path, current, errors.Join(ErrUnsupported, ErrNestedRepository))
		}
	}
	return nil
}
