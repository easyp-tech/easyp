// Package gitsnapshot exposes one immutable Git commit as a filesystem.
package gitsnapshot

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// ErrGitlink marks a tracked submodule boundary without reading its contents.
var ErrGitlink = errors.New("gitlink boundary")

// FS reads tracked trees and blobs without consulting a checkout or host links.
// Symlinks are exposed to fs.ReadLinkFS; callers choose their bounded resolution policy.
type FS struct {
	repo  *git.Repository
	tree  *object.Tree
	raw   *repositoryFS
	mu    sync.Mutex
	trees map[plumbing.Hash]indexedTree
	sizes map[plumbing.Hash]int64
}

type indexedTree struct {
	entries []object.TreeEntry
	byName  map[string]object.TreeEntry
}

func indexTree(tree *object.Tree) indexedTree {
	index := indexedTree{entries: tree.Entries, byName: make(map[string]object.TreeEntry, len(tree.Entries))}
	for _, entry := range tree.Entries {
		index.byName[entry.Name] = entry
	}
	return index
}

// New opens the tree pinned by commit in repo.
func New(repo *git.Repository, commit plumbing.Hash) (*FS, error) {
	revision, err := repo.CommitObject(commit)
	if err != nil {
		return nil, fmt.Errorf("CommitObject: %w", err)
	}
	tree, err := revision.Tree()
	if err != nil {
		return nil, fmt.Errorf("Tree: %w", err)
	}
	return &FS{
		repo: repo, tree: tree,
		trees: map[plumbing.Hash]indexedTree{tree.Hash: indexTree(tree)},
		sizes: make(map[plumbing.Hash]int64),
	}, nil
}

// Immutable objects are decoded once per pinned filesystem. Keeping both the
// directory index and blob metadata avoids reopening packs for every path prefix.
func (f *FS) indexedTree(hash plumbing.Hash) (indexedTree, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if tree, ok := f.trees[hash]; ok {
		return tree, nil
	}
	tree, err := f.repo.TreeObject(hash)
	if err != nil {
		return indexedTree{}, fmt.Errorf("TreeObject: %w", err)
	}
	index := indexTree(tree)
	f.trees[hash] = index
	return index, nil
}

func (f *FS) blobSize(hash plumbing.Hash) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if size, ok := f.sizes[hash]; ok {
		return size, nil
	}
	blob, err := f.repo.BlobObject(hash)
	if err != nil {
		return 0, fmt.Errorf("BlobObject: %w", err)
	}
	// A go-git Blob retains its decoded body. Keep only the size here so metadata
	// traversal cannot retain the full repository outside the storage's bounded LRU.
	f.sizes[hash] = blob.Size
	return blob.Size, nil
}

func (f *FS) lookup(name string) (object.TreeEntry, error) {
	if !fs.ValidPath(name) {
		return object.TreeEntry{}, fs.ErrInvalid
	}
	if name == "." {
		return object.TreeEntry{Name: ".", Mode: filemode.Dir, Hash: f.tree.Hash}, nil
	}
	tree, err := f.indexedTree(f.tree.Hash)
	if err != nil {
		return object.TreeEntry{}, fmt.Errorf("indexedTree: %w", err)
	}
	parts := strings.Split(name, "/")
	for i, part := range parts {
		entry, ok := tree.byName[part]
		if !ok {
			return object.TreeEntry{}, fs.ErrNotExist
		}
		if entry.Mode == filemode.Submodule {
			return object.TreeEntry{}, fmt.Errorf("%w at %q", ErrGitlink, strings.Join(parts[:i+1], "/"))
		}
		if i == len(parts)-1 {
			return entry, nil
		}
		if entry.Mode != filemode.Dir {
			return object.TreeEntry{}, fmt.Errorf("non-directory Git component %q", strings.Join(parts[:i+1], "/"))
		}
		child, err := f.indexedTree(entry.Hash)
		if err != nil {
			return object.TreeEntry{}, fmt.Errorf("indexedTree: %w", err)
		}
		tree = child
	}
	return object.TreeEntry{}, fs.ErrNotExist
}

func (f *FS) info(entry object.TreeEntry) (fileInfo, error) {
	info := fileInfo{name: entry.Name}
	switch entry.Mode {
	case filemode.Dir:
		info.mode = fs.ModeDir | 0o755
	case filemode.Regular, filemode.Deprecated:
		info.mode = 0o644
	case filemode.Executable:
		info.mode = 0o755
	case filemode.Symlink:
		info.mode = fs.ModeSymlink | 0o777
	case filemode.Submodule:
		return fileInfo{}, fmt.Errorf("%w at %q", ErrGitlink, entry.Name)
	default:
		return fileInfo{}, fmt.Errorf("unsupported Git mode %s at %q", entry.Mode, entry.Name)
	}
	if entry.Mode != filemode.Dir {
		size, err := f.blobSize(entry.Hash)
		if err != nil {
			return fileInfo{}, fmt.Errorf("blobSize: %w", err)
		}
		info.size = size
	}
	return info, nil
}

// Lstat describes a tracked node without following a symbolic link.
func (f *FS) Lstat(name string) (fs.FileInfo, error) {
	if f.raw != nil {
		return f.raw.Lstat(name)
	}
	entry, err := f.lookup(name)
	if err != nil {
		return nil, &fs.PathError{Op: "lstat", Path: name, Err: err}
	}
	info, err := f.info(entry)
	if err != nil {
		return nil, &fs.PathError{Op: "lstat", Path: name, Err: err}
	}
	return info, nil
}

// ReadLink reads the committed pointer bytes of a tracked symbolic link.
func (f *FS) ReadLink(name string) (string, error) {
	if f.raw != nil {
		return f.raw.ReadLink(name)
	}
	entry, err := f.lookup(name)
	if err != nil {
		return "", &fs.PathError{Op: "readlink", Path: name, Err: err}
	}
	if entry.Mode != filemode.Symlink {
		return "", &fs.PathError{Op: "readlink", Path: name, Err: fs.ErrInvalid}
	}
	file, err := f.Open(name)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(file)
	closeErr := file.Close()
	if err != nil {
		return "", fmt.Errorf("ReadAll: %w", err)
	}
	if closeErr != nil {
		return "", fmt.Errorf("Close: %w", closeErr)
	}
	return string(data), nil
}

// Open reads a committed blob or directory. It never follows links itself.
func (f *FS) Open(name string) (fs.File, error) {
	if f.raw != nil {
		return f.raw.Open(name)
	}
	entry, err := f.lookup(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	info, err := f.info(entry)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	if info.IsDir() {
		entries, err := f.ReadDir(name)
		if err != nil {
			return nil, err
		}
		return &directory{info: info, entries: entries}, nil
	}
	blob, err := f.repo.BlobObject(entry.Hash)
	if err != nil {
		return nil, fmt.Errorf("BlobObject: %w", err)
	}
	reader, err := blob.Reader()
	if err != nil {
		return nil, fmt.Errorf("Reader: %w", err)
	}
	return &blobFile{ReadCloser: reader, info: info}, nil
}

// ReadDir enumerates direct tracked children in lexical order.
func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	if f.raw != nil {
		return f.raw.ReadDir(name)
	}
	entry, err := f.lookup(name)
	if err != nil {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}
	if entry.Mode != filemode.Dir {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}
	tree, err := f.indexedTree(entry.Hash)
	if err != nil {
		return nil, fmt.Errorf("indexedTree: %w", err)
	}
	entries := make([]fs.DirEntry, 0, len(tree.entries))
	for _, child := range tree.entries {
		// Gitlinks are represented as inaccessible directories so bounded walkers
		// can omit auxiliary submodules but fail when a selected path crosses one.
		if child.Mode == filemode.Submodule {
			entries = append(entries, gitlinkEntry{name: child.Name})
			continue
		}
		info, err := f.info(child)
		if err != nil {
			return nil, fmt.Errorf("info: %w", err)
		}
		entries = append(entries, fs.FileInfoToDirEntry(info))
	}
	slices.SortFunc(entries, func(a, b fs.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	return entries, nil
}

type fileInfo struct {
	name string
	mode fs.FileMode
	size int64
}

func (i fileInfo) Name() string       { return path.Base(i.name) }
func (i fileInfo) Size() int64        { return i.size }
func (i fileInfo) Mode() fs.FileMode  { return i.mode }
func (i fileInfo) ModTime() time.Time { return time.Time{} }
func (i fileInfo) IsDir() bool        { return i.mode.IsDir() }
func (i fileInfo) Sys() any           { return nil }

type blobFile struct {
	io.ReadCloser
	info fileInfo
}

func (f *blobFile) Stat() (fs.FileInfo, error) { return f.info, nil }

type directory struct {
	info    fileInfo
	entries []fs.DirEntry
	offset  int
}

func (d *directory) Stat() (fs.FileInfo, error) { return d.info, nil }
func (d *directory) Close() error               { return nil }
func (d *directory) Read([]byte) (int, error)   { return 0, fs.ErrInvalid }
func (d *directory) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.offset >= len(d.entries) && n > 0 {
		return nil, io.EOF
	}
	end := len(d.entries)
	if n > 0 {
		end = min(end, d.offset+n)
	}
	result := d.entries[d.offset:end]
	d.offset = end
	return result, nil
}

type gitlinkEntry struct{ name string }

func (e gitlinkEntry) Name() string      { return e.name }
func (e gitlinkEntry) IsDir() bool       { return true }
func (e gitlinkEntry) Type() fs.FileMode { return fs.ModeDir }
func (e gitlinkEntry) Info() (fs.FileInfo, error) {
	return nil, fmt.Errorf("%w at %q", ErrGitlink, e.name)
}
