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
	"strconv"
	"strings"

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
}

// NewRepository reads an immutable tree through Git's format-aware object
// commands. In particular, SHA-256 object IDs must not become SHA-1 hashes.
func NewRepository(ctx context.Context, directory, commit string) (*FS, error) {
	if !objectID(commit) {
		return nil, fmt.Errorf("invalid pinned Git commit %q", commit)
	}
	backend := &repositoryFS{ctx: ctx, directory: directory, entries: map[string]repositoryEntry{
		".": {info: fileInfo{name: ".", mode: fs.ModeDir | 0o755}},
	}}
	raw, err := backend.command(nil, "ls-tree", "-rzt", "--full-tree", commit)
	if err != nil {
		return nil, fmt.Errorf("command: %w", err)
	}
	objects := make(map[string]bool)
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
		if fields[1] == "blob" {
			objects[entry.object] = true
		}
		backend.entries[name] = entry
	}
	ids := make([]string, 0, len(objects))
	for id := range objects {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	if len(ids) != 0 {
		input := strings.NewReader(strings.Join(ids, "\n") + "\n")
		checked, err := backend.command(input, "cat-file", "--batch-check")
		if err != nil {
			return nil, fmt.Errorf("command: %w", err)
		}
		sizes := make(map[string]int64, len(ids))
		for line := range strings.SplitSeq(strings.TrimSpace(string(checked)), "\n") {
			fields := strings.Fields(line)
			if len(fields) != 3 || fields[1] != "blob" {
				return nil, fmt.Errorf("invalid Git blob metadata %q", line)
			}
			size, err := strconv.ParseInt(fields[2], 10, 64)
			if err != nil || size < 0 {
				return nil, fmt.Errorf("invalid Git blob size %q", line)
			}
			sizes[fields[0]] = size
		}
		for name, entry := range backend.entries {
			entry.info.size = sizes[entry.object]
			backend.entries[name] = entry
		}
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
	return entry, nil
}

func (f *repositoryFS) Lstat(name string) (fs.FileInfo, error) {
	entry, err := f.lookup(name)
	if err != nil {
		return nil, &fs.PathError{Op: "lstat", Path: name, Err: err}
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
	raw, err := f.command(nil, "cat-file", "blob", entry.object)
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
	raw, err := f.command(nil, "cat-file", "blob", entry.object)
	if err != nil {
		return nil, fmt.Errorf("command: %w", err)
	}
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
			children = append(children, fs.FileInfoToDirEntry(entry.info))
		}
	}
	slices.SortFunc(children, func(a, b fs.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	return children, nil
}
