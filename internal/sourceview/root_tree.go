package sourceview

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
)

// rootTree serves an os.Root by slash-separated name through cached handles
// for its directories, so a lookup costs one system call instead of one per
// path component. Every handle is an os.Root opened from its parent's handle,
// which confines it to that directory; View only asks for directories it has
// already resolved to physical, link-free paths.
type rootTree struct {
	root *os.Root

	mu   sync.Mutex
	dirs map[string]*os.Root
}

func newRootTree(root *os.Root) *rootTree {
	return &rootTree{root: root, dirs: map[string]*os.Root{".": root}}
}

// dir returns the handle for a directory, opening missing ancestors.
func (t *rootTree) dir(name string) (*os.Root, error) {
	t.mu.Lock()
	handle, ok := t.dirs[name]
	t.mu.Unlock()
	if ok {
		return handle, nil
	}
	parent, err := t.dir(path.Dir(name))
	if err != nil {
		return nil, err
	}
	handle, err = parent.OpenRoot(path.Base(name))
	if err != nil {
		return nil, err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if existing, ok := t.dirs[name]; ok {
		return existing, handle.Close()
	}
	t.dirs[name] = handle
	return handle, nil
}

// split returns the directory handle and base name for name.
func (t *rootTree) split(op, name string) (*os.Root, string, error) {
	if !fs.ValidPath(name) {
		return nil, "", &fs.PathError{Op: op, Path: name, Err: fs.ErrInvalid}
	}
	if name == "." {
		return t.root, ".", nil
	}
	handle, err := t.dir(path.Dir(name))
	if err != nil {
		return nil, "", pathError(op, name, err)
	}
	return handle, path.Base(name), nil
}

func (t *rootTree) Open(name string) (fs.File, error) {
	handle, base, err := t.split("open", name)
	if err != nil {
		return nil, err
	}
	file, err := handle.Open(base)
	if err != nil {
		return nil, pathError("open", name, err)
	}
	return file, nil
}

func (t *rootTree) Lstat(name string) (fs.FileInfo, error) {
	handle, base, err := t.split("lstat", name)
	if err != nil {
		return nil, err
	}
	info, err := handle.Lstat(base)
	if err != nil {
		return nil, pathError("lstat", name, err)
	}
	return info, nil
}

func (t *rootTree) ReadLink(name string) (string, error) {
	handle, base, err := t.split("readlink", name)
	if err != nil {
		return "", err
	}
	target, err := handle.Readlink(base)
	if err != nil {
		return "", pathError("readlink", name, err)
	}
	return target, nil
}

func (t *rootTree) ReadDir(name string) ([]fs.DirEntry, error) {
	file, err := t.Open(name)
	if err != nil {
		return nil, err
	}
	dir, ok := file.(fs.ReadDirFile)
	if !ok {
		return nil, errors.Join(&fs.PathError{Op: "readdir", Path: name, Err: errors.New("not implemented")}, file.Close())
	}
	entries, err := dir.ReadDir(-1)
	slices.SortFunc(entries, func(a, b fs.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	return entries, errors.Join(err, file.Close())
}

// Close closes every directory handle, the root included.
func (t *rootTree) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	var errs []error
	for name, handle := range t.dirs {
		if name != "." {
			errs = append(errs, handle.Close())
		}
	}
	t.dirs = nil
	errs = append(errs, t.root.Close())
	return errors.Join(errs...)
}

// pathError reports the full name instead of the handle-relative base.
func pathError(op, name string, err error) error {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return &fs.PathError{Op: op, Path: name, Err: pathErr.Err}
	}
	return &fs.PathError{Op: op, Path: name, Err: err}
}
