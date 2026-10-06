package sourceview

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
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
	resolved, err := view.Resolve(ctx, filepath.ToSlash(logical))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("Resolve: %w", localResolutionError(logical, resolved, err)), root.Close())
	}
	if err := CheckLocalResolution(root, resolved); err != nil {
		permitted := errors.Is(err, ErrNestedRepository) && len(nestedAllowed) > 0 && nestedAllowed[0] != nil && nestedAllowed[0](resolved)
		if !permitted {
			return nil, errors.Join(err, root.Close())
		}
	}
	if !resolved.Info.Mode().IsRegular() {
		return nil, errors.Join(ErrUnsupported, root.Close())
	}
	if selected != nil && !selected(resolved) {
		return nil, errors.Join(&fs.PathError{Op: "open", Path: logical, Err: fs.ErrNotExist}, root.Close())
	}
	file, err := root.Open(filepath.FromSlash(resolved.Path))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("Open: %w", err), root.Close())
	}
	opened, err := file.Stat()
	if err != nil {
		return nil, errors.Join(fmt.Errorf("Stat: %w", err), file.Close(), root.Close())
	}
	if !os.SameFile(resolved.Info, opened) || resolved.Info.Mode() != opened.Mode() || resolved.Info.Size() != opened.Size() || !resolved.Info.ModTime().Equal(opened.ModTime()) {
		return nil, errors.Join(ErrChanged, file.Close(), root.Close())
	}
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(err, file.Close(), root.Close())
	}
	return &localFile{File: file, root: root}, nil
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

func localView(rootPath string) (*os.Root, *View, error) {
	canonical, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return nil, nil, fmt.Errorf("EvalSymlinks: %w", err)
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return nil, nil, fmt.Errorf("Abs: %w", err)
	}
	root, err := os.OpenRoot(canonical)
	if err != nil {
		return nil, nil, fmt.Errorf("OpenRoot: %w", err)
	}
	requested, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("Abs: %w", err), root.Close())
	}
	original, err := os.Stat(requested)
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("Stat: %w", err), root.Close())
	}
	opened, err := root.Stat(".")
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("Stat: %w", err), root.Close())
	}
	if !os.SameFile(original, opened) {
		return nil, nil, errors.Join(ErrChanged, root.Close())
	}
	view, err := NewLocal(root.FS(), canonical, requested)
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("NewLocal: %w", err), root.Close())
	}
	return root, view, nil
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
	if len(resolved.Links) == 0 {
		return nil
	}
	directories := []string{filepath.Dir(filepath.FromSlash(resolved.Path))}
	if resolved.Info != nil && resolved.Info.IsDir() {
		directories[0] = filepath.FromSlash(resolved.Path)
	}
	for _, link := range resolved.Links {
		directories = append(directories, filepath.Dir(filepath.FromSlash(link.Path)))
	}
	seen := make(map[string]bool)
	for _, directory := range directories {
		for current := directory; current != "."; current = filepath.Dir(current) {
			if seen[current] {
				break
			}
			seen[current] = true
			_, err := root.Lstat(filepath.Join(current, ".git"))
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
