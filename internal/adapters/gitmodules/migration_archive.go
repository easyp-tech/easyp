package gitmodules

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"golang.org/x/mod/sumdb/dirhash"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

// hashMigrationProtoArchive preserves the regular-file proof helper. Production
// migration uses both evidenced installer policies through migrationArchiveHashes.
func hashMigrationProtoArchive(ctx context.Context, checkout string, files []string) (string, error) {
	roots, err := readMigrationLegacyRoots(checkout, files)
	if err != nil {
		return "", fmt.Errorf("readMigrationLegacyRoots: %w", err)
	}
	selected := make(map[string][]byte)
	for _, name := range files {
		if path.Ext(name) != ".proto" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(checkout, filepath.FromSlash(name)))
		if err != nil {
			return "", fmt.Errorf("ReadFile: %w", err)
		}
		selected[renameMigrationLegacyFile(name, roots)] = data
	}
	nodes, err := readMigrationProtoArchive(ctx, checkout, "HEAD", files)
	if err != nil {
		return "", fmt.Errorf("readMigrationProtoArchive: %w", err)
	}
	hashes, err := migrationArchiveHashes(ctx, nodes, roots, selected)
	if err != nil {
		return "", fmt.Errorf("migrationArchiveHashes: %w", err)
	}
	return hashes[0], nil
}

// readMigrationProtoArchive reads only committed ZIP entries. Link payloads are
// pointer strings; they are never extracted onto, or resolved against, the host.
// Git reproduces the released *.proto pathspec and export-ignore/export-subst.
func readMigrationProtoArchive(ctx context.Context, checkout, commit string, files []string) (nodes []migrationArchiveNode, resultErr error) {
	return readMigrationArchive(ctx, checkout, commit, files, nil, true)
}

func readMigrationSourceArchive(ctx context.Context, checkout, commit string, files, roots []string) ([]migrationArchiveNode, error) {
	return readMigrationArchive(ctx, checkout, commit, files, roots, false)
}

func readMigrationArchive(ctx context.Context, checkout, commit string, files, roots []string, fullDigest bool) (nodes []migrationArchiveNode, resultErr error) {
	directory, err := os.MkdirTemp("", "easyp-migration-archive-")
	if err != nil {
		return nil, fmt.Errorf("MkdirTemp: %w", err)
	}
	defer func() {
		removeErr := os.RemoveAll(directory)
		if removeErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("RemoveAll: %w", removeErr))
		}
	}()
	archivePath := filepath.Join(directory, "protos.zip")
	_, err = gitV1(ctx, checkout, "archive", "--format=zip", "--output="+archivePath, commit, "--", "*.proto")
	if err != nil {
		return nil, fmt.Errorf("gitV1: %w", err)
	}
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("OpenReader: %w", err)
	}
	defer func() {
		closeErr := archive.Close()
		if closeErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("Close: %w", closeErr))
		}
	}()
	tracked := make(map[string]bool, len(files))
	for _, name := range files {
		tracked[name] = true
	}
	for _, file := range archive.File {
		err = ctx.Err()
		if err != nil {
			return nil, fmt.Errorf("Err: %w", err)
		}
		name := strings.TrimSuffix(file.Name, "/")
		if !fs.ValidPath(name) || path.Clean(name) != name || strings.ContainsAny(name, "\\\x00") {
			return nil, fmt.Errorf("unsupported archive path %q", file.Name)
		}
		mode := file.Mode()
		if !mode.IsDir() && !mode.IsRegular() && mode&fs.ModeSymlink == 0 {
			return nil, fmt.Errorf("unsupported non-regular archive file %q", name)
		}
		if !mode.IsDir() && !tracked[name] {
			return nil, fmt.Errorf("unsupported untracked archive file %q", name)
		}
		var data []byte
		if !mode.IsDir() && (fullDigest || path.Ext(name) == ".proto" || mode&fs.ModeSymlink != 0) {
			data, err = readMigrationArchiveFile(file)
			if err != nil {
				return nil, fmt.Errorf("readMigrationArchiveFile: %w", err)
			}
		}
		nodes = append(nodes, migrationArchiveNode{name: file.Name, mode: mode, data: data})
	}
	if !fullDigest {
		if err := loadMigrationArchiveAliasTargets(ctx, archive.File, nodes, roots); err != nil {
			return nil, fmt.Errorf("loadMigrationArchiveAliasTargets: %w", err)
		}
	}
	return nodes, nil
}

// A proto alias can point to an archived file with another extension. Load only
// those target bodies needed under an evidenced installer namespace.
func loadMigrationArchiveAliasTargets(ctx context.Context, files []*zip.File, nodes []migrationArchiveNode, roots []string) error {
	indices := make(map[string]int, len(nodes))
	archives := make(map[string]*zip.File, len(files))
	for i, node := range nodes {
		indices[node.name] = i
	}
	for _, file := range files {
		archives[file.Name] = file
	}
	for _, rewrite := range []bool{false, true} {
		installed, _, err := installMigrationArchive(nodes, roots, rewrite)
		if err != nil {
			continue
		}
		originals := make(map[string]string, len(nodes))
		for _, node := range nodes {
			originals[renameMigrationLegacyFile(node.name, roots)] = node.name
		}
		view := sourceview.New(installed)
		for _, node := range nodes {
			if node.mode.IsDir() || path.Ext(node.name) != ".proto" {
				continue
			}
			resolved, err := view.Resolve(ctx, renameMigrationLegacyFile(node.name, roots))
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				continue
			}
			original := originals[resolved.Path]
			index, exists := indices[original]
			if !exists || !nodes[index].mode.IsRegular() || nodes[index].data != nil {
				continue
			}
			data, err := readMigrationArchiveFile(archives[original])
			if err != nil {
				return fmt.Errorf("readMigrationArchiveFile: %w", err)
			}
			nodes[index].data = data
		}
	}
	return nil
}

// migrationArchiveHashes independently reconstructs released installer layouts:
// v0.15 renamed node paths and retained pointer strings; v0.16/v0.17 additionally
// renamed each pointer's lexically resolved target. A successful candidate must
// preserve the complete proto import namespace and bytes of the current view.
func migrationArchiveHashes(ctx context.Context, nodes []migrationArchiveNode, roots []string, selected map[string][]byte) ([]string, error) {
	var hashes []string
	var failures []error
	for _, rewrite := range []bool{false, true} {
		installed, names, err := installMigrationArchive(nodes, roots, rewrite)
		if err != nil {
			failures = append(failures, fmt.Errorf("installMigrationArchive: %w", err))
			continue
		}
		view := sourceview.New(installed)
		legacy := make(map[string][]byte)
		for _, name := range names {
			file, openErr := view.Open(ctx, name)
			if openErr != nil {
				err = fmt.Errorf("Open: %w", openErr)
				break
			}
			data, readErr := io.ReadAll(file)
			closeErr := file.Close()
			err = errors.Join(readErr, closeErr)
			if err != nil {
				err = fmt.Errorf("ReadAll: %w", err)
				break
			}
			if path.Ext(name) == ".proto" {
				legacy[name] = data
			}
		}
		if err == nil {
			err = validateMigrationSourceContents("legacy archive", legacy, selected)
		}
		if err != nil {
			failures = append(failures, err)
			continue
		}
		hash, err := dirhash.Hash1(names, func(name string) (io.ReadCloser, error) {
			file, err := view.Open(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("Open: %w", err)
			}
			return file, nil
		})
		if err != nil {
			failures = append(failures, fmt.Errorf("Hash1: %w", err))
			continue
		}
		if !slices.Contains(hashes, hash) {
			hashes = append(hashes, hash)
		}
	}
	if len(hashes) == 0 {
		return nil, errors.Join(failures...)
	}
	return hashes, nil
}

func installMigrationArchive(nodes []migrationArchiveNode, roots []string, rewrite bool) (migrationArchiveFS, []string, error) {
	installed := migrationArchiveFS{".": {name: ".", mode: fs.ModeDir}}
	originals := make(map[string]string)
	directories := make(map[string]bool)
	for _, original := range nodes {
		// Retain trailing slashes while renaming directory nodes, as extract.Archive did.
		destination := path.Clean(renameMigrationLegacyFile(original.name, roots))
		if original.mode.IsDir() {
			for directory := destination; directory != "."; directory = path.Dir(directory) {
				directories[directory] = true
			}
			continue
		}
		if other, exists := originals[destination]; exists {
			return nil, nil, fmt.Errorf("legacy file collision at %q between %q and %q", destination, other, original.name)
		}
		node := original
		node.name = destination
		if rewrite && node.mode&fs.ModeSymlink != 0 {
			target, err := rewriteMigrationArchiveTarget(original.name, string(node.data), destination, roots)
			if err != nil {
				return nil, nil, fmt.Errorf("rewriteMigrationArchiveTarget: %w", err)
			}
			node.data = []byte(target)
		}
		installed[destination], originals[destination] = node, original.name
		for directory := path.Dir(destination); directory != "."; directory = path.Dir(directory) {
			directories[directory] = true
		}
	}
	names := make([]string, 0, len(originals))
	for name := range originals {
		if directories[name] {
			return nil, nil, fmt.Errorf("legacy directory/file collision at %q", name)
		}
		names = append(names, name)
	}
	for directory := range directories {
		installed[directory] = migrationArchiveNode{name: directory, mode: fs.ModeDir}
	}
	slices.Sort(names)
	return installed, names, nil
}

func rewriteMigrationArchiveTarget(original, target, destination string, roots []string) (string, error) {
	if strings.ContainsAny(target, "\\\x00") || strings.HasPrefix(target, "//") || (len(target) >= 2 && target[1] == ':') {
		return "", sourceview.ErrUnsupported
	}
	if path.IsAbs(target) {
		return "", sourceview.ErrOutsideRoot
	}
	resolved := path.Join(path.Dir(original), target)
	if !fs.ValidPath(resolved) {
		return "", sourceview.ErrOutsideRoot
	}
	renamed := path.Clean(renameMigrationLegacyFile(resolved, roots))
	relative, err := filepath.Rel(filepath.FromSlash(path.Dir(destination)), filepath.FromSlash(renamed))
	if err != nil {
		return "", fmt.Errorf("Rel: %w", err)
	}
	return filepath.ToSlash(relative), nil
}

// validateMigrationArchiveRegularSources guards archive attributes during new
// acquisition without requiring an old installer to support a logical alias.
func validateMigrationArchiveRegularSources(nodes []migrationArchiveNode, snapshot string, files []string) error {
	archived := make(map[string]migrationArchiveNode, len(nodes))
	for _, node := range nodes {
		archived[node.name] = node
	}
	for _, name := range files {
		if path.Ext(name) != ".proto" {
			continue
		}
		node, exists := archived[name]
		if !exists {
			return fmt.Errorf("legacy archive source selection omits %q; manual migration is required", name)
		}
		current, err := os.ReadFile(filepath.Join(snapshot, filepath.FromSlash(name)))
		if err != nil {
			return fmt.Errorf("ReadFile: %w", err)
		}
		if !bytes.Equal(node.data, current) {
			return fmt.Errorf("legacy archive source selection changes %q contents; manual migration is required", name)
		}
	}
	return nil
}

func readMigrationArchiveFile(file *zip.File) (content []byte, resultErr error) {
	reader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("Open: %w", err)
	}
	defer func() {
		closeErr := reader.Close()
		if closeErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("Close: %w", closeErr))
		}
	}()
	content, err = io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %w", err)
	}
	return content, nil
}

// migrationArchiveFS contains only immutable installed ZIP nodes. Open never
// follows links; sourceview is the only resolver and cannot leave this map.
type migrationArchiveFS map[string]migrationArchiveNode

func (f migrationArchiveFS) Lstat(name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "Lstat", Path: name, Err: fs.ErrInvalid}
	}
	node, exists := f[name]
	if !exists {
		return nil, &fs.PathError{Op: "Lstat", Path: name, Err: fs.ErrNotExist}
	}
	return node, nil
}

func (f migrationArchiveFS) ReadLink(name string) (string, error) {
	info, err := f.Lstat(name)
	if err != nil {
		return "", fmt.Errorf("Lstat: %w", err)
	}
	if info.Mode()&fs.ModeSymlink == 0 {
		return "", &fs.PathError{Op: "ReadLink", Path: name, Err: fs.ErrInvalid}
	}
	return string(f[name].data), nil
}

func (f migrationArchiveFS) Open(name string) (fs.File, error) {
	info, err := f.Lstat(name)
	if err != nil {
		return nil, fmt.Errorf("Lstat: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, &fs.PathError{Op: "Open", Path: name, Err: sourceview.ErrUnsupported}
	}
	return migrationArchiveFile{Reader: bytes.NewReader(f[name].data), info: info}, nil
}

type migrationArchiveNode struct {
	name string
	mode fs.FileMode
	data []byte
}

func (n migrationArchiveNode) Name() string       { return path.Base(n.name) }
func (n migrationArchiveNode) Size() int64        { return int64(len(n.data)) }
func (n migrationArchiveNode) Mode() fs.FileMode  { return n.mode }
func (n migrationArchiveNode) ModTime() time.Time { return time.Time{} }
func (n migrationArchiveNode) IsDir() bool        { return n.mode.IsDir() }
func (n migrationArchiveNode) Sys() any           { return nil }

type migrationArchiveFile struct {
	*bytes.Reader
	info fs.FileInfo
}

func (f migrationArchiveFile) Stat() (fs.FileInfo, error) { return f.info, nil }
func (migrationArchiveFile) Close() error                 { return nil }
