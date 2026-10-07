package modules

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

type resolvedFileState struct {
	name       string
	data       []byte
	exists     bool
	resolution sourceview.Resolution
	allowed    SourceRoots
}

type resolvedFileChange struct {
	name string
	data []byte
	mode os.FileMode
}

// resolvedFilesTransaction extends the module writer with bounded, staged
// sources and metadata. Multiple renames are not crash atomic. State rechecks
// detect concurrent changes before commit, but cannot exclude a noncooperating
// writer racing the final check and rename.
type resolvedFilesTransaction struct {
	root           *os.Root
	requestedRoot  string
	canonicalRoot  string
	rootInfo       os.FileInfo
	expected       map[string]resolvedFileState
	parents        map[string]os.FileInfo
	changes        map[string]resolvedFileChange
	inputs         []*resolvedFilesTransaction
	validateInputs func() error
	beforeCommit   func() error
	rename         func(*os.Root, string, string) error
	stage          func(*os.Root, string, []byte, os.FileMode) error
}

func newResolvedFilesTransaction(directory string) (_ *resolvedFilesTransaction, resultErr error) {
	requested, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("Abs: %w", err)
	}
	canonical, err := filepath.EvalSymlinks(requested)
	if err != nil {
		return nil, fmt.Errorf("EvalSymlinks: %w", err)
	}
	info, err := os.Stat(requested)
	if err != nil {
		return nil, fmt.Errorf("Stat: %w", err)
	}
	root, err := os.OpenRoot(canonical)
	if err != nil {
		return nil, fmt.Errorf("OpenRoot: %w", err)
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, root.Close())
		}
	}()
	tx := &resolvedFilesTransaction{
		root: root, requestedRoot: requested, canonicalRoot: canonical, rootInfo: info,
		expected: make(map[string]resolvedFileState), parents: make(map[string]os.FileInfo),
		changes: make(map[string]resolvedFileChange), stage: stageResolvedFile,
		rename: func(root *os.Root, before, after string) error { return root.Rename(before, after) },
	}
	err = tx.checkRoot()
	if err != nil {
		return nil, fmt.Errorf("checkRoot: %w", err)
	}
	return tx, nil
}

func (tx *resolvedFilesTransaction) close() error {
	var failures []error
	for _, input := range tx.inputs {
		failures = append(failures, input.close())
	}
	err := tx.root.Close()
	if err != nil {
		failures = append(failures, fmt.Errorf("Close: %w", err))
	}
	return errors.Join(failures...)
}

func (tx *resolvedFilesTransaction) checkRoot() error {
	canonical, err := filepath.EvalSymlinks(tx.requestedRoot)
	if err != nil {
		return fmt.Errorf("EvalSymlinks: %w", err)
	}
	info, err := os.Stat(tx.requestedRoot)
	if err != nil {
		return fmt.Errorf("Stat: %w", err)
	}
	opened, err := tx.root.Stat(".")
	if err != nil {
		return fmt.Errorf("Stat: %w", err)
	}
	if canonical != tx.canonicalRoot || !os.SameFile(tx.rootInfo, info) || !os.SameFile(tx.rootInfo, opened) || tx.rootInfo.Mode() != info.Mode() {
		return fmt.Errorf("module directory %q changed: %w", tx.requestedRoot, sourceview.ErrChanged)
	}
	return nil
}

func (tx *resolvedFilesTransaction) capture(name string, allowed SourceRoots) (resolvedFileState, error) {
	current, err := tx.readState(name, allowed)
	if err != nil {
		return resolvedFileState{}, fmt.Errorf("readState: %w", err)
	}
	if before, exists := tx.expected[name]; exists {
		if !sameResolvedFile(before, current) {
			return resolvedFileState{}, fmt.Errorf("file %q changed since planning: %w", name, sourceview.ErrChanged)
		}
		return before, nil
	}
	err = tx.captureParents(current.resolution.Path)
	if err != nil {
		return resolvedFileState{}, fmt.Errorf("captureParents: %w", err)
	}
	tx.expected[name] = current
	return current, nil
}

func (tx *resolvedFilesTransaction) readState(name string, allowed SourceRoots) (_ resolvedFileState, resultErr error) {
	if !filepath.IsLocal(name) || name == "." || filepath.Clean(name) != name || strings.Contains(name, "\\") {
		return resolvedFileState{}, fmt.Errorf("unsafe relative file name %q", name)
	}
	view, err := sourceview.NewLocal(tx.root.FS(), tx.canonicalRoot, tx.requestedRoot)
	if err != nil {
		return resolvedFileState{}, fmt.Errorf("NewLocal: %w", err)
	}
	resolved, err := view.Resolve(context.Background(), filepath.ToSlash(name))
	if errors.Is(err, os.ErrNotExist) && len(resolved.Links) == 0 {
		return resolvedFileState{name: name, resolution: sourceview.Resolution{Path: filepath.ToSlash(name)}, allowed: allowed}, nil
	}
	if err != nil {
		return resolvedFileState{}, fmt.Errorf("Resolve: %w", err)
	}
	err = sourceview.CheckLocalResolution(tx.root, resolved)
	physical := filepath.Join(tx.canonicalRoot, filepath.FromSlash(resolved.Path))
	if errors.Is(err, sourceview.ErrNestedRepository) && allowed.ownsSelectedPhysical(physical) {
		err = nil
	}
	if err != nil {
		return resolvedFileState{}, fmt.Errorf("CheckLocalResolution: %w", err)
	}
	if !resolved.Info.Mode().IsRegular() {
		return resolvedFileState{}, fmt.Errorf("file %q must resolve to a regular file: %w", name, sourceview.ErrUnsupported)
	}
	if len(allowed) > 0 && !allowed.ownsSelectedPhysical(physical) {
		return resolvedFileState{}, fmt.Errorf("source %q resolves outside its owning module roots: %w", name, sourceview.ErrOutsideRoot)
	}
	file, err := tx.root.Open(filepath.FromSlash(resolved.Path))
	if err != nil {
		return resolvedFileState{}, fmt.Errorf("Open: %w", err)
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("Close: %w", closeErr))
		}
	}()
	info, err := file.Stat()
	if err != nil {
		return resolvedFileState{}, fmt.Errorf("Stat: %w", err)
	}
	if !sameResolvedInfo(resolved.Info, info) {
		return resolvedFileState{}, fmt.Errorf("file %q changed while opening: %w", name, sourceview.ErrChanged)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return resolvedFileState{}, fmt.Errorf("ReadAll: %w", err)
	}
	after, err := view.Resolve(context.Background(), filepath.ToSlash(name))
	if err != nil {
		return resolvedFileState{}, fmt.Errorf("Resolve: %w", err)
	}
	if !sameResolvedPath(resolved, after) || int64(len(data)) != info.Size() {
		return resolvedFileState{}, fmt.Errorf("file %q changed while reading: %w", name, sourceview.ErrChanged)
	}
	return resolvedFileState{name: name, data: data, exists: true, resolution: resolved, allowed: allowed}, nil
}

func sameResolvedInfo(before, after os.FileInfo) bool {
	return before != nil && after != nil && os.SameFile(before, after) && before.Mode() == after.Mode() && before.Size() == after.Size() && before.ModTime().Equal(after.ModTime())
}

func sameResolvedPath(before, after sourceview.Resolution) bool {
	if before.Path != after.Path || len(before.Links) != len(after.Links) || !sameResolvedInfo(before.Info, after.Info) {
		return false
	}
	for i, link := range before.Links {
		other := after.Links[i]
		if link.Path != other.Path || link.Target != other.Target || !sameResolvedInfo(link.Info, other.Info) {
			return false
		}
	}
	return true
}

func sameResolvedFile(before, after resolvedFileState) bool {
	return before.exists == after.exists && (!before.exists || (sameResolvedPath(before.resolution, after.resolution) && bytes.Equal(before.data, after.data)))
}

func (tx *resolvedFilesTransaction) captureParents(name string) error {
	for parent := filepath.Dir(filepath.FromSlash(name)); parent != "."; parent = filepath.Dir(parent) {
		info, err := tx.root.Lstat(parent)
		if err != nil {
			return fmt.Errorf("Lstat: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("destination parent %q must be a regular directory: %w", parent, sourceview.ErrUnsupported)
		}
		if before, exists := tx.parents[parent]; exists && (!os.SameFile(before, info) || before.Mode() != info.Mode()) {
			return fmt.Errorf("parent %q changed since planning: %w", parent, sourceview.ErrChanged)
		}
		tx.parents[parent] = info
	}
	return nil
}

func (tx *resolvedFilesTransaction) plan(name string, data []byte, defaultMode os.FileMode) error {
	before, exists := tx.expected[name]
	if !exists {
		return fmt.Errorf("destination %q was not captured during planning", name)
	}
	if before.exists && bytes.Equal(before.data, data) {
		return nil
	}
	mode := defaultMode
	if before.exists {
		mode = before.resolution.Info.Mode()
	}
	target := filepath.FromSlash(before.resolution.Path)
	if existing, duplicate := tx.changes[target]; duplicate && !bytes.Equal(existing.data, data) {
		return fmt.Errorf("conflicting writes to physical destination %q", target)
	}
	tx.changes[target] = resolvedFileChange{name: target, data: bytes.Clone(data), mode: mode}
	return nil
}

func (tx *resolvedFilesTransaction) verify() error {
	err := tx.checkRoot()
	if err != nil {
		return fmt.Errorf("checkRoot: %w", err)
	}
	for _, parent := range slices.Sorted(maps.Keys(tx.parents)) {
		info, err := tx.root.Lstat(parent)
		if err != nil {
			return fmt.Errorf("Lstat: %w", err)
		}
		if before := tx.parents[parent]; !os.SameFile(before, info) || before.Mode() != info.Mode() {
			return fmt.Errorf("parent %q changed since planning: %w", parent, sourceview.ErrChanged)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(tx.expected)) {
		before := tx.expected[name]
		current, err := tx.readState(name, before.allowed)
		if err != nil {
			return fmt.Errorf("readState: %w", err)
		}
		if !sameResolvedFile(before, current) {
			return fmt.Errorf("file %q changed since planning: %w", name, sourceview.ErrChanged)
		}
	}
	for _, input := range tx.inputs {
		err = input.verify()
		if err != nil {
			return fmt.Errorf("verify: %w", err)
		}
	}
	if tx.validateInputs != nil {
		err = tx.validateInputs()
		if err != nil {
			return fmt.Errorf("validateInputs: %w", err)
		}
	}
	return nil
}

type stagedResolvedFile struct {
	change    resolvedFileChange
	output    string
	backup    string
	before    resolvedFileState
	installed resolvedFileState
}

func (tx *resolvedFilesTransaction) apply() (resultErr error) {
	err := tx.verify()
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	if len(tx.changes) == 0 {
		return nil
	}
	temporary := ".easyp-resolved-" + rand.Text()
	err = tx.root.Mkdir(temporary, 0o700)
	if err != nil {
		return fmt.Errorf("Mkdir: %w", err)
	}
	keepRecovery := false
	defer func() {
		if !keepRecovery {
			err := tx.root.RemoveAll(temporary)
			if err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("RemoveAll: %w", err))
			}
		}
	}()
	var staged []stagedResolvedFile
	for _, name := range slices.Sorted(maps.Keys(tx.changes)) {
		change := tx.changes[name]
		before, err := tx.readState(name, nil)
		if err != nil {
			return fmt.Errorf("readState: %w", err)
		}
		item := stagedResolvedFile{change: change, before: before,
			output: filepath.Join(temporary, fmt.Sprintf("output-%d", len(staged))),
			backup: filepath.Join(temporary, fmt.Sprintf("original-%d", len(staged))),
		}
		err = tx.stage(tx.root, item.output, change.data, change.mode)
		if err != nil {
			return fmt.Errorf("stage: %w", err)
		}
		if before.exists {
			err = tx.stage(tx.root, item.backup, before.data, before.resolution.Info.Mode())
			if err != nil {
				return fmt.Errorf("stage: %w", err)
			}
		}
		item.installed, err = tx.readState(item.output, nil)
		if err != nil {
			return fmt.Errorf("readState: %w", err)
		}
		item.installed.name, item.installed.resolution.Path = change.name, filepath.ToSlash(change.name)
		staged = append(staged, item)
	}
	if tx.beforeCommit != nil {
		err = tx.beforeCommit()
		if err != nil {
			return fmt.Errorf("beforeCommit: %w", err)
		}
	}
	err = tx.verify()
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	for index, item := range staged {
		current, checkErr := tx.readState(item.change.name, nil)
		if checkErr == nil && !sameResolvedFile(item.before, current) {
			checkErr = fmt.Errorf("file %q changed before commit: %w", item.change.name, sourceview.ErrChanged)
		}
		err = checkErr
		if err == nil {
			if item.before.exists {
				err = tx.rename(tx.root, item.output, item.change.name)
			} else {
				// Link refuses a concurrently created destination, including a
				// dangling alias. Staging cleanup removes the extra link.
				err = tx.root.Link(item.output, item.change.name)
			}
		}
		if err == nil {
			continue
		}
		rollbackErr := tx.rollback(staged[:index])
		if rollbackErr != nil {
			keepRecovery = true
			rollbackErr = fmt.Errorf("rollback: %w; recovery copies remain in %q", rollbackErr, filepath.Join(tx.canonicalRoot, temporary))
		}
		return errors.Join(fmt.Errorf("install: %w", err), rollbackErr)
	}
	return nil
}

func (tx *resolvedFilesTransaction) rollback(applied []stagedResolvedFile) error {
	var failures []error
	for index := len(applied) - 1; index >= 0; index-- {
		item := applied[index]
		current, err := tx.readState(item.change.name, nil)
		if err == nil && !sameResolvedFile(item.installed, current) {
			err = fmt.Errorf("file %q changed after commit: %w", item.change.name, sourceview.ErrChanged)
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("readState: %w", err))
			continue
		}
		if item.before.exists {
			err = tx.rename(tx.root, item.backup, item.change.name)
		} else {
			err = tx.root.Remove(item.change.name)
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("restore: %w", err))
		}
	}
	return errors.Join(failures...)
}

func stageResolvedFile(root *os.Root, name string, data []byte, mode os.FileMode) (resultErr error) {
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("OpenFile: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("Close: %w", err))
		}
	}()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("Write: %w", err)
	}
	err = file.Chmod(mode)
	if err != nil {
		return fmt.Errorf("Chmod: %w", err)
	}
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("Sync: %w", err)
	}
	return nil
}
