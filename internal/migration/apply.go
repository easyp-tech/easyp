package migration

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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

type snapshot struct {
	name     string
	data     []byte
	mode     os.FileMode
	exists   bool
	info     os.FileInfo
	input    bool
	resolved string
	links    []sourceview.Link
	pointer  string
}

type fileChange struct {
	name    string
	content []byte
	mode    os.FileMode
}

// transaction stages every output and rollback copy before modifying destinations.
// Ordinary write failures restore the previous contents and modes. This is not a
// crash-atomic multi-file commit: process termination or machine failure can leave
// a partially applied migration. Rechecks detect changes visible before a write,
// but cannot exclude noncooperating writers between the final check and rename.
// Root-relative operations contain writes even if a path is concurrently moved.
type transaction struct {
	root          string
	expected      map[string]snapshot
	changes       []fileChange
	candidates    []Output
	requestedRoot string
	rootInfo      os.FileInfo
	beforeApply   func() error
	rename        func(*os.Root, string, string) error
	link          func(*os.Root, string, string) error
	stage         func(*os.Root, string, []byte, os.FileMode) error
}

type stagedChange struct {
	change    fileChange
	output    string
	rollback  string
	installed snapshot
}

func newTransaction(root string) (*transaction, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("Abs: %w", err)
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, fmt.Errorf("Lstat: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("migration root %q must be a directory, not a symlink", abs)
	}
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("EvalSymlinks: %w", err)
	}
	tx := &transaction{
		root:          canonical,
		requestedRoot: abs,
		rootInfo:      info,
		expected:      make(map[string]snapshot),
		rename:        func(root *os.Root, oldName, newName string) error { return root.Rename(oldName, newName) },
		link:          func(root *os.Root, oldName, newName string) error { return root.Link(oldName, newName) },
		stage:         stageFile,
	}
	err = tx.checkRoot()
	if err != nil {
		return nil, fmt.Errorf("checkRoot: %w", err)
	}
	return tx, nil
}

func (t *transaction) checkRoot() error {
	info, err := os.Lstat(t.requestedRoot)
	if err != nil {
		return fmt.Errorf("Lstat: %w", err)
	}
	if !info.IsDir() || !os.SameFile(t.rootInfo, info) || t.rootInfo.Mode() != info.Mode() {
		return fmt.Errorf("migration root %q changed since planning", t.requestedRoot)
	}
	canonical, err := filepath.EvalSymlinks(t.requestedRoot)
	if err != nil {
		return fmt.Errorf("EvalSymlinks: %w", err)
	}
	if canonical != t.root {
		return fmt.Errorf("migration root %q resolves differently since planning", t.requestedRoot)
	}
	info, err = os.Lstat(t.root)
	if err != nil {
		return fmt.Errorf("Lstat: %w", err)
	}
	if !info.IsDir() || !os.SameFile(t.rootInfo, info) {
		return fmt.Errorf("migration root %q changed since planning", t.root)
	}
	return nil
}

func (t *transaction) openRoot() (_ *os.Root, resultErr error) {
	err := t.checkRoot()
	if err != nil {
		return nil, fmt.Errorf("checkRoot: %w", err)
	}
	root, err := os.OpenRoot(t.requestedRoot)
	if err != nil {
		return nil, fmt.Errorf("OpenRoot: %w", err)
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, closeRoot(root))
		}
	}()
	info, err := root.Stat(".")
	if err != nil {
		return nil, fmt.Errorf("Stat: %w", err)
	}
	if !os.SameFile(t.rootInfo, info) {
		return nil, fmt.Errorf("migration root %q changed while opening", t.root)
	}
	return root, nil
}

func (t *transaction) capture(name string) (_ snapshot, resultErr error) {
	if previous, ok := t.expected[name]; ok && previous.input {
		return t.captureInput(name)
	}
	root, err := t.openRoot()
	if err != nil {
		return snapshot{}, fmt.Errorf("openRoot: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, closeRoot(root)) }()
	current, err := readSnapshot(root, name)
	if err != nil {
		return snapshot{}, fmt.Errorf("readSnapshot: %w", err)
	}
	if previous, ok := t.expected[name]; ok {
		if !sameSnapshot(previous, current) {
			return snapshot{}, fmt.Errorf("file %q changed since planning", name)
		}
		return previous, nil
	}
	t.expected[name] = current
	return current, nil
}

func (t *transaction) captureInput(name string) (_ snapshot, resultErr error) {
	root, err := t.openRoot()
	if err != nil {
		return snapshot{}, fmt.Errorf("openRoot: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, closeRoot(root)) }()
	current, err := readInputSnapshot(root, name)
	if err != nil {
		return snapshot{}, fmt.Errorf("readInputSnapshot: %w", err)
	}
	if previous, ok := t.expected[name]; ok {
		if !sameSnapshot(previous, current) {
			return snapshot{}, fmt.Errorf("file %q changed since planning", name)
		}
		return previous, nil
	}
	t.expected[name] = current
	return current, nil
}

func readInputSnapshot(root *os.Root, name string) (_ snapshot, resultErr error) {
	if err := checkRelativeName(name); err != nil {
		return snapshot{}, fmt.Errorf("checkRelativeName: %w", err)
	}
	canonical, err := filepath.EvalSymlinks(root.Name())
	if err != nil {
		return snapshot{}, fmt.Errorf("EvalSymlinks: %w", err)
	}
	view, err := sourceview.NewLocal(root.FS(), canonical, root.Name())
	if err != nil {
		return snapshot{}, fmt.Errorf("NewLocal: %w", err)
	}
	resolved, err := view.Resolve(context.Background(), filepath.ToSlash(name))
	if errors.Is(err, os.ErrNotExist) && len(resolved.Links) == 0 {
		return snapshot{name: name, input: true}, nil
	}
	if err != nil {
		return snapshot{}, fmt.Errorf("Resolve: %w", err)
	}
	if err := sourceview.CheckLocalResolution(root, resolved); err != nil {
		return snapshot{}, fmt.Errorf("CheckLocalResolution: %w", err)
	}
	if !resolved.Info.Mode().IsRegular() {
		return snapshot{}, fmt.Errorf("input %q must resolve to a regular file", name)
	}
	current, err := readSnapshot(root, filepath.FromSlash(resolved.Path))
	if err != nil {
		return snapshot{}, fmt.Errorf("readSnapshot: %w", err)
	}
	current.name, current.input, current.resolved, current.links = name, true, resolved.Path, resolved.Links
	leaf, err := root.Lstat(name)
	if err != nil {
		return snapshot{}, fmt.Errorf("Lstat: %w", err)
	}
	if leaf.Mode()&os.ModeSymlink != 0 {
		current.pointer, err = root.Readlink(name)
		if err != nil {
			return snapshot{}, fmt.Errorf("Readlink: %w", err)
		}
	}
	after, err := view.Resolve(context.Background(), filepath.ToSlash(name))
	if err != nil {
		return snapshot{}, fmt.Errorf("Resolve: %w", err)
	}
	if !sameResolution(resolved, after) {
		return snapshot{}, fmt.Errorf("input %q changed while reading", name)
	}
	return current, nil
}

func sameResolution(before, after sourceview.Resolution) bool {
	if before.Path != after.Path || len(before.Links) != len(after.Links) {
		return false
	}
	if !sameFileState(before.Info, after.Info) {
		return false
	}
	for i, link := range before.Links {
		other := after.Links[i]
		if link.Path != other.Path || link.Target != other.Target || !sameFileState(link.Info, other.Info) {
			return false
		}
	}
	return true
}

func readSnapshot(root *os.Root, name string) (_ snapshot, resultErr error) {
	err := checkRelativeName(name)
	if err != nil {
		return snapshot{}, fmt.Errorf("checkRelativeName: %w", err)
	}
	err = checkParents(root, name)
	if err != nil {
		return snapshot{}, fmt.Errorf("checkParents: %w", err)
	}
	info, err := root.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return snapshot{name: name}, nil
	}
	if err != nil {
		return snapshot{}, fmt.Errorf("Lstat: %w", err)
	}
	if !info.Mode().IsRegular() {
		return snapshot{}, fmt.Errorf("file %q must be a regular file, not a symlink or directory", name)
	}
	file, err := root.Open(name)
	if err != nil {
		return snapshot{}, fmt.Errorf("Open: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("Close: %w", err))
		}
	}()
	opened, err := file.Stat()
	if err != nil {
		return snapshot{}, fmt.Errorf("Stat: %w", err)
	}
	if !sameFileState(info, opened) {
		return snapshot{}, fmt.Errorf("file %q changed while opening", name)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return snapshot{}, fmt.Errorf("ReadAll: %w", err)
	}
	after, err := root.Lstat(name)
	if err != nil {
		return snapshot{}, fmt.Errorf("Lstat: %w", err)
	}
	if !sameFileState(info, after) || int64(len(data)) != after.Size() {
		return snapshot{}, fmt.Errorf("file %q changed while reading", name)
	}
	err = checkParents(root, name)
	if err != nil {
		return snapshot{}, fmt.Errorf("checkParents: %w", err)
	}
	return snapshot{name: name, data: data, mode: info.Mode(), exists: true, info: info}, nil
}

func sameFileState(before, after os.FileInfo) bool {
	return os.SameFile(before, after) && before.Mode() == after.Mode() &&
		before.Size() == after.Size() && before.ModTime().Equal(after.ModTime())
}

func sameSnapshot(before, after snapshot) bool {
	if before.exists != after.exists {
		return false
	}
	if before.resolved != after.resolved || before.pointer != after.pointer || len(before.links) != len(after.links) {
		return false
	}
	for i, link := range before.links {
		other := after.links[i]
		if link.Path != other.Path || link.Target != other.Target || !sameFileState(link.Info, other.Info) {
			return false
		}
	}
	return !before.exists || (os.SameFile(before.info, after.info) && before.mode == after.mode &&
		sha256.Sum256(before.data) == sha256.Sum256(after.data))
}

func checkRelativeName(name string) error {
	if !filepath.IsLocal(name) || name == "." || filepath.Clean(name) != name || strings.Contains(name, "\\") {
		return fmt.Errorf("unsafe relative file name %q", name)
	}
	return nil
}

func checkParents(root *os.Root, name string) error {
	parent := filepath.Dir(name)
	if parent == "." {
		return nil
	}
	current := ""
	for _, component := range strings.Split(parent, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := root.Lstat(current)
		if err != nil {
			return fmt.Errorf("Lstat: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("parent %q must be a directory, not a symlink", current)
		}
	}
	return nil
}

func (t *transaction) verify(root *os.Root) error {
	err := t.checkRoot()
	if err != nil {
		return fmt.Errorf("checkRoot: %w", err)
	}
	for _, name := range slices.Sorted(maps.Keys(t.expected)) {
		err = recheckSnapshot(root, t.expected[name])
		if err != nil {
			return fmt.Errorf("recheckSnapshot: %w", err)
		}
	}
	return nil
}

func recheckSnapshot(root *os.Root, expected snapshot) error {
	var current snapshot
	var err error
	if expected.input {
		current, err = readInputSnapshot(root, expected.name)
	} else {
		current, err = readSnapshot(root, expected.name)
	}
	if err != nil {
		return fmt.Errorf("readSnapshot: %w", err)
	}
	if !sameSnapshot(expected, current) {
		return fmt.Errorf("file %q changed since planning", expected.name)
	}
	return nil
}

func (t *transaction) apply() (resultErr error) {
	root, err := t.openRoot()
	if err != nil {
		return fmt.Errorf("openRoot: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, closeRoot(root)) }()
	err = t.validateChanges()
	if err != nil {
		return fmt.Errorf("validateChanges: %w", err)
	}
	err = t.verify(root)
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	if len(t.changes) == 0 {
		return nil
	}
	temporary := ".easyp-migrate-" + rand.Text()
	err = root.Mkdir(temporary, 0o700)
	if err != nil {
		return fmt.Errorf("Mkdir: %w", err)
	}
	keepTemporary := false
	defer func() {
		if keepTemporary {
			return
		}
		err := root.RemoveAll(temporary)
		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("RemoveAll: %w", err))
		}
	}()
	staged, err := t.stageChanges(root, temporary)
	if err != nil {
		return fmt.Errorf("stageChanges: %w", err)
	}
	if t.beforeApply != nil {
		err = t.beforeApply()
		if err != nil {
			return fmt.Errorf("beforeApply: %w", err)
		}
	}
	// Recheck every source and destination after potentially slow staging, before
	// making any externally visible change.
	err = t.verify(root)
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	for index, item := range staged {
		err = t.install(root, item)
		if err == nil {
			continue
		}
		rollbackErr := t.rollback(root, staged[:index])
		if rollbackErr != nil {
			// Keep recovery copies if the filesystem also refuses rollback.
			keepTemporary = true
			rollbackErr = fmt.Errorf("rollback: %w; recovery copies remain in %q", rollbackErr, filepath.Join(t.root, temporary))
		}
		return errors.Join(fmt.Errorf("install: %w", err), rollbackErr)
	}
	return nil
}

func (t *transaction) validateChanges() error {
	seen := make(map[string]bool)
	for _, change := range t.changes {
		err := checkRelativeName(change.name)
		if err != nil {
			return fmt.Errorf("checkRelativeName: %w", err)
		}
		if filepath.Dir(change.name) != "." {
			return fmt.Errorf("destination %q must be directly inside the migration root", change.name)
		}
		if _, ok := t.expected[change.name]; !ok {
			return fmt.Errorf("destination %q was not captured during planning", change.name)
		}
		for _, input := range t.expected {
			if len(input.links) > 0 && input.resolved == filepath.ToSlash(change.name) {
				return fmt.Errorf("destination %q overlaps linked input target", change.name)
			}
			for _, link := range input.links {
				if link.Path == filepath.ToSlash(change.name) && input.name != change.name {
					return fmt.Errorf("destination %q overlaps linked input hop for %q", change.name, input.name)
				}
			}
		}
		if seen[change.name] {
			return fmt.Errorf("duplicate destination %q", change.name)
		}
		seen[change.name] = true
		if !change.mode.IsRegular() {
			return fmt.Errorf("destination %q has a nonregular file mode", change.name)
		}
	}
	return nil
}

func (t *transaction) stageChanges(root *os.Root, temporary string) ([]stagedChange, error) {
	staged := make([]stagedChange, 0, len(t.changes))
	for index, change := range t.changes {
		item := stagedChange{
			change:   change,
			output:   filepath.Join(temporary, fmt.Sprintf("output-%d", index)),
			rollback: filepath.Join(temporary, fmt.Sprintf("original-%d", index)),
		}
		err := t.stage(root, item.output, change.content, change.mode)
		if err != nil {
			return nil, fmt.Errorf("stage: %w", err)
		}
		info, err := root.Lstat(item.output)
		if err != nil {
			return nil, fmt.Errorf("Lstat: %w", err)
		}
		item.installed = snapshot{name: change.name, data: change.content, mode: change.mode, exists: true, info: info}
		before := t.expected[change.name]
		if before.exists {
			if before.pointer != "" {
				err = root.Symlink(before.pointer, item.rollback)
			} else {
				err = t.stage(root, item.rollback, before.data, before.mode)
			}
			if err != nil {
				return nil, fmt.Errorf("stage: %w", err)
			}
		}
		staged = append(staged, item)
	}
	return staged, nil
}

func (t *transaction) install(root *os.Root, item stagedChange) error {
	err := t.checkRoot()
	if err != nil {
		return fmt.Errorf("checkRoot: %w", err)
	}
	before := t.expected[item.change.name]
	err = recheckSnapshot(root, before)
	if err != nil {
		return fmt.Errorf("recheckSnapshot: %w", err)
	}
	if before.exists {
		err = t.rename(root, item.output, item.change.name)
		if err != nil {
			return fmt.Errorf("rename: %w", err)
		}
		return nil
	}
	// Link atomically refuses an existing destination, including dangling links.
	// The staging link is removed with the temporary directory after completion.
	err = t.link(root, item.output, item.change.name)
	if err != nil {
		return fmt.Errorf("link: %w", err)
	}
	return nil
}

func (t *transaction) rollback(root *os.Root, applied []stagedChange) error {
	var failures []error
	for index := len(applied) - 1; index >= 0; index-- {
		item := applied[index]
		err := recheckSnapshot(root, item.installed)
		if err != nil {
			failures = append(failures, fmt.Errorf("recheckSnapshot: %w", err))
			continue
		}
		if t.expected[item.change.name].exists {
			err = t.rename(root, item.rollback, item.change.name)
			if err != nil {
				failures = append(failures, fmt.Errorf("rename: %w", err))
			}
			continue
		}
		err = root.Remove(item.change.name)
		if err != nil {
			failures = append(failures, fmt.Errorf("Remove: %w", err))
		}
	}
	return errors.Join(failures...)
}

func stageFile(root *os.Root, name string, data []byte, mode os.FileMode) (resultErr error) {
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
	// Chmod restores exact permissions independently of the process umask.
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

func closeRoot(root *os.Root) error {
	err := root.Close()
	if err != nil {
		return fmt.Errorf("Close: %w", err)
	}
	return nil
}
