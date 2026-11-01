package modules

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

// Resolution belongs to a logical import, not to a particular cached directory.
// Equal copies keep the ordinary root-order precedence and canonical owner.
type tidyImportResolution struct {
	source     v1ImportSource
	root       SourceRoot
	candidates []v1ImportSource
}

func (view *tidySourceView) importCandidates(ctx context.Context, name string) (tidyImportResolution, error) {
	resolution := tidyImportResolution{}
	// Re-probe used names against the filesystem after cache verification.
	// An immutable inventory must not hide a newly added import candidate.
	roots := slices.Clone(view.roots)
	for index := range roots {
		roots[index].index = nil
	}
	allowed := roots.FileAllowed()
	seen := make(map[string]bool)
	for _, root := range roots {
		err := ctx.Err()
		if err != nil {
			return tidyImportResolution{}, fmt.Errorf("Err: %w", err)
		}
		candidate, err := resolveV1ImportInRoot(name, root, roots, allowed)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return tidyImportResolution{}, fmt.Errorf("resolveV1ImportInRoot: %w", err)
		}
		if seen[candidate.path] {
			continue
		}
		seen[candidate.path] = true
		if len(resolution.candidates) == 0 {
			resolution.source, resolution.root = candidate, root
		}
		resolution.candidates = append(resolution.candidates, candidate)
	}
	if len(resolution.candidates) > 0 {
		return resolution, nil
	}
	source, err := resolveV1BuiltinImport(name)
	if err != nil {
		return tidyImportResolution{}, fmt.Errorf("resolveV1BuiltinImport: %w", err)
	}
	resolution.source = source
	return resolution, nil
}

func (view *tidySourceView) resolveImport(ctx context.Context, name string) (tidyImportResolution, error) {
	resolution, err := view.importCandidates(ctx, name)
	if err != nil {
		return tidyImportResolution{}, fmt.Errorf("importCandidates: %w", err)
	}
	var canonical []byte
	for offset, candidate := range resolution.candidates {
		data, err := view.readSourceFile(candidate.path)
		if err != nil {
			return tidyImportResolution{}, fmt.Errorf("readSourceFile: %w", err)
		}
		if offset == 0 {
			canonical = data
			continue
		}
		if !bytes.Equal(canonical, data) {
			return tidyImportResolution{}, fmt.Errorf("duplicate import path %q: %s and %s have different contents", name, resolution.source.path, candidate.path)
		}
	}
	view.mu.Lock()
	defer view.mu.Unlock()
	if previous, exists := view.imports[name]; exists && !slices.Equal(previous.candidates, resolution.candidates) {
		return tidyImportResolution{}, fmt.Errorf("import candidates for %q changed: %w", name, sourceview.ErrChanged)
	}
	view.imports[name] = resolution
	return resolution, nil
}

// Re-probe used names so adding, removing or excluding another candidate cannot
// invalidate deduplication after comparison. File observations check all bytes
// and alias identities, including noncanonical copies, separately.
func (view *tidySourceView) verifyImportCandidates(ctx context.Context) error {
	view.mu.Lock()
	expected := make(map[string]tidyImportResolution, len(view.imports))
	for name, resolution := range view.imports {
		expected[name] = resolution
	}
	view.mu.Unlock()
	names := make([]string, 0, len(expected))
	for name := range expected {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		current, err := view.importCandidates(ctx, name)
		if err != nil {
			return fmt.Errorf("importCandidates: %s: %w", name, err)
		}
		if !slices.Equal(expected[name].candidates, current.candidates) {
			return fmt.Errorf("import candidates for %q changed: %w", name, sourceview.ErrChanged)
		}
	}
	return nil
}

func (view *tidySourceView) validateConsumerTargets(ctx context.Context, files []string) error {
	consumer := view.roots[0].Module
	for _, name := range files {
		path := filepath.Join(view.observations.requestedRoot, name)
		for _, root := range view.roots {
			if root.Module != consumer || !sourcePathWithin(path, root.Path) || !root.allows(path, root.Path) {
				continue
			}
			importPath, err := filepath.Rel(root.Path, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			_, err = view.resolveImport(ctx, filepath.ToSlash(importPath))
			if err != nil {
				return fmt.Errorf("resolveImport: %s: %w", name, err)
			}
		}
	}
	return nil
}
