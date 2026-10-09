package gitsnapshot

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/easyp-tech/easyp/internal/adapters/gitcommand"
)

type repositoryEntry struct {
	object string
	info   fileInfo
	link   bool
}

type repositoryFS struct {
	ctx       context.Context
	directory string
	entries   map[string]repositoryEntry
	mu        sync.Mutex
	sizes     map[string]int64
	trees     map[string]map[string]string
}

// NewRepository reads an immutable tree through Git's format-aware object
// commands. In particular, SHA-256 object IDs must not become SHA-1 hashes.
func NewRepository(ctx context.Context, directory, commit string) (*FS, error) {
	if !objectID(commit) {
		return nil, fmt.Errorf("invalid pinned Git commit %q", commit)
	}
	commit = strings.ToLower(commit)
	backend := &repositoryFS{ctx: ctx, directory: directory, sizes: make(map[string]int64), trees: make(map[string]map[string]string), entries: map[string]repositoryEntry{
		".": {info: fileInfo{name: ".", mode: fs.ModeDir | 0o755}},
	}}
	commitData, err := backend.object("commit", commit)
	if err != nil {
		return nil, fmt.Errorf("object: %w", err)
	}
	treeLine, _, _ := strings.Cut(string(commitData), "\n")
	rootTree := strings.TrimPrefix(treeLine, "tree ")
	if !objectID(rootTree) {
		return nil, fmt.Errorf("invalid commit tree")
	}
	root := backend.entries["."]
	root.object = rootTree
	backend.entries["."] = root
	raw, err := backend.command(nil, "ls-tree", "-rzt", "--full-tree", commit)
	if err != nil {
		return nil, fmt.Errorf("command: %w", err)
	}
	for record := range strings.SplitSeq(string(raw), "\x00") {
		if record == "" {
			continue
		}
		metadata, name, ok := strings.Cut(record, "\t")
		fields := strings.Fields(metadata)
		if !ok || len(fields) != 3 || !fs.ValidPath(name) || !objectID(fields[2]) {
			return nil, fmt.Errorf("invalid Git tree entry %q", record)
		}
		entry := repositoryEntry{object: fields[2], info: fileInfo{name: path.Base(name)}}
		switch fields[0] {
		case "040000":
			entry.info.mode = fs.ModeDir | 0o755
		case "100644":
			entry.info.mode = 0o644
		case "100755":
			entry.info.mode = 0o755
		case "120000":
			entry.info.mode = fs.ModeSymlink | 0o777
		case "160000":
			entry.link = true
		default:
			return nil, fmt.Errorf("unsupported Git tree mode %q", fields[0])
		}
		backend.entries[name] = entry
	}
	return &FS{raw: backend}, nil
}

func objectID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func (f *repositoryFS) command(stdin io.Reader, args ...string) ([]byte, error) {
	raw, err := gitcommand.Run(f.ctx, f.directory, stdin, args...)
	if err != nil {
		return nil, fmt.Errorf("Run: %w", err)
	}
	return raw, nil
}

func (f *repositoryFS) lookup(name string) (repositoryEntry, error) {
	if !fs.ValidPath(name) {
		return repositoryEntry{}, fs.ErrInvalid
	}
	for prefix := name; prefix != "."; prefix = path.Dir(prefix) {
		if entry, found := f.entries[prefix]; found && entry.link {
			return repositoryEntry{}, fmt.Errorf("%w at %q", ErrGitlink, prefix)
		}
	}
	entry, found := f.entries[name]
	if !found {
		return repositoryEntry{}, fs.ErrNotExist
	}
	if err := f.verifyPath(name); err != nil {
		return repositoryEntry{}, err
	}
	return entry, nil
}

func (f *repositoryFS) Lstat(name string) (fs.FileInfo, error) {
	entry, err := f.lookup(name)
	if err != nil {
		return nil, &fs.PathError{Op: "lstat", Path: name, Err: err}
	}
	if !entry.info.IsDir() {
		size, err := f.objectSize(entry.object)
		if err != nil {
			return nil, fmt.Errorf("objectSize: %w", err)
		}
		entry.info.size = size
	}
	return entry.info, nil
}

func (f *repositoryFS) ReadLink(name string) (string, error) {
	entry, err := f.lookup(name)
	if err != nil {
		return "", &fs.PathError{Op: "readlink", Path: name, Err: err}
	}
	if entry.info.mode&fs.ModeSymlink == 0 {
		return "", &fs.PathError{Op: "readlink", Path: name, Err: fs.ErrInvalid}
	}
	raw, err := f.object("blob", entry.object)
	if err != nil {
		return "", fmt.Errorf("command: %w", err)
	}
	return string(raw), nil
}

func (f *repositoryFS) Open(name string) (fs.File, error) {
	entry, err := f.lookup(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	if entry.info.IsDir() {
		children, err := f.ReadDir(name)
		if err != nil {
			return nil, fmt.Errorf("ReadDir: %w", err)
		}
		return &directory{info: entry.info, entries: children}, nil
	}
	raw, err := f.object("blob", entry.object)
	if err != nil {
		return nil, fmt.Errorf("command: %w", err)
	}
	entry.info.size = int64(len(raw))
	return &blobFile{ReadCloser: io.NopCloser(bytes.NewReader(raw)), info: entry.info}, nil
}

func (f *repositoryFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entry, err := f.lookup(name)
	if err != nil || !entry.info.IsDir() {
		if err == nil {
			err = fs.ErrInvalid
		}
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}
	var children []fs.DirEntry
	for child, entry := range f.entries {
		if child == "." || path.Dir(child) != name {
			continue
		}
		if entry.link {
			children = append(children, gitlinkEntry{name: path.Base(child)})
		} else {
			children = append(children, repositoryDirEntry{fsys: f, name: child, info: entry.info})
		}
	}
	slices.SortFunc(children, func(a, b fs.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	return children, nil
}

type repositoryDirEntry struct {
	fsys *repositoryFS
	name string
	info fileInfo
}

func (e repositoryDirEntry) Name() string               { return e.info.Name() }
func (e repositoryDirEntry) IsDir() bool                { return e.info.IsDir() }
func (e repositoryDirEntry) Type() fs.FileMode          { return e.info.Mode().Type() }
func (e repositoryDirEntry) Info() (fs.FileInfo, error) { return e.fsys.Lstat(e.name) }
