package modules

import (
	"bytes"
	"context"
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

// resolvedFileObservation exposes owned bytes and the captured physical path.
// File identities, aliases and bounded root checks stay with the transaction.
type resolvedFileObservation struct {
	data   []byte
	exists bool
	path   string
}

// resolvedFileObservations owns bounded file states and read-only child inputs.
// A child can observe and recheck files, but cannot propose or commit writes.
type resolvedFileObservations struct {
	root          *os.Root
	requestedRoot string
	canonicalRoot string
	rootInfo      os.FileInfo
	expected      map[string]resolvedFileState
	parents       map[string]os.FileInfo
	inputs        []*resolvedFileObservations
}

func newResolvedFileObservations(directory string) (_ *resolvedFileObservations, resultErr error) {
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
	tx := &resolvedFileObservations{
		root: root, requestedRoot: requested, canonicalRoot: canonical, rootInfo: info,
		expected: make(map[string]resolvedFileState), parents: make(map[string]os.FileInfo),
	}
	err = tx.checkRoot()
	if err != nil {
		return nil, fmt.Errorf("checkRoot: %w", err)
	}
	return tx, nil
}

func (tx *resolvedFileObservations) close() error {
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

func (tx *resolvedFileObservations) checkRoot() error {
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

func (tx *resolvedFileObservations) capture(name string, allowed SourceRoots) (resolvedFileObservation, error) {
	current, err := tx.readState(name, allowed)
	if err != nil {
		return resolvedFileObservation{}, fmt.Errorf("readState: %w", err)
	}
	if before, exists := tx.expected[name]; exists {
		if !sameResolvedFile(before, current) {
			return resolvedFileObservation{}, fmt.Errorf("file %q changed since planning: %w", name, sourceview.ErrChanged)
		}
		return observeResolvedFile(before), nil
	}
	err = tx.captureParents(current.resolution.Path)
	if err != nil {
		return resolvedFileObservation{}, fmt.Errorf("captureParents: %w", err)
	}
	tx.expected[name] = current
	return observeResolvedFile(current), nil
}

func (tx *resolvedFileObservations) readState(name string, allowed SourceRoots) (_ resolvedFileState, resultErr error) {
	if !filepath.IsLocal(name) || name == "." || filepath.Clean(name) != name || strings.Contains(name, "\\") {
		return resolvedFileState{}, fmt.Errorf("unsafe relative file name %q", name)
	}
	view, err := sourceview.NewLocal(tx.root.FS(), tx.canonicalRoot, tx.requestedRoot)
	if err != nil {
		return resolvedFileState{}, fmt.Errorf("NewLocal: %w", err)
	}
	resolved, err := view.Resolve(context.Background(), filepath.ToSlash(name))
	if errors.Is(err, os.ErrNotExist) && len(resolved.Links) == 0 {
		return resolvedFileState{name: name, resolution: sourceview.Resolution{Path: filepath.ToSlash(name)}, allowed: slices.Clone(allowed)}, nil
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
	return resolvedFileState{name: name, data: data, exists: true, resolution: resolved, allowed: slices.Clone(allowed)}, nil
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

func (tx *resolvedFileObservations) captureParents(name string) error {
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

func observeResolvedFile(state resolvedFileState) resolvedFileObservation {
	return resolvedFileObservation{data: bytes.Clone(state.data), exists: state.exists, path: state.resolution.Path}
}

func (tx *resolvedFileObservations) capturedFile(name string) (resolvedFileObservation, bool) {
	state, captured := tx.expected[name]
	return observeResolvedFile(state), captured
}

func (tx *resolvedFileObservations) capturedTarget(name string) (string, bool) {
	state, captured := tx.expected[name]
	return state.resolution.Path, captured
}

func (tx *resolvedFileObservations) observeDirectory(directory string) (*resolvedFileObservations, error) {
	input, err := newResolvedFileObservations(directory)
	if err != nil {
		return nil, fmt.Errorf("newResolvedFileObservations: %w", err)
	}
	tx.inputs = append(tx.inputs, input)
	return input, nil
}

func (tx *resolvedFileObservations) verifyCapturedInputs() error {
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
		err = input.verifyCapturedInputs()
		if err != nil {
			return fmt.Errorf("verifyCapturedInputs: %w", err)
		}
	}
	return nil
}
