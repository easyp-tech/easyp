package modules

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/bufbuild/protocompile"
	compilerimports "github.com/bufbuild/protocompile/wellknownimports"

	"github.com/easyp-tech/easyp/wellknownimports"
)

// tidySourceView reads the proposed consumer bytes and captures unchanged
// reachable inputs for the same pre-commit state check as the project writer.
type tidySourceView struct {
	observations *resolvedFileObservations
	roots        SourceRoots
	proposed     map[string][]byte
	scopes       map[string]*resolvedFileObservations
	pinned       map[string]map[string][]byte
	old          []tidyOldNamespace
	mu           sync.Mutex
}

func newTidySourceView(observations *resolvedFileObservations, roots SourceRoots) *tidySourceView {
	return &tidySourceView{
		observations: observations, roots: roots,
		proposed: make(map[string][]byte), scopes: make(map[string]*resolvedFileObservations),
		pinned: make(map[string]map[string][]byte),
	}
}

func (view *tidySourceView) propose(target string, data []byte) {
	view.mu.Lock()
	defer view.mu.Unlock()
	view.proposed[target] = bytes.Clone(data)
}

func (view *tidySourceView) retainPinnedSources(resolved graphResolveResult) {
	for _, entry := range resolved.lockFile().Modules {
		fetched, exists := resolved.currentRevision(entry)
		if exists && fetched.Inspection != nil {
			view.pinned[entry.Source] = make(map[string][]byte)
			for _, file := range fetched.Inspection.Files {
				view.pinned[entry.Source][file.Path] = file.Content
			}
		}
	}
}

func (view *tidySourceView) readSourceFile(path string) ([]byte, error) {
	view.mu.Lock()
	defer view.mu.Unlock()
	for _, root := range view.roots {
		if !sourcePathWithin(path, root.Path) || !root.allows(path, root.Path) {
			continue
		}
		tx := view.observations
		boundary := root.boundary()
		if !sourcePathWithin(path, tx.requestedRoot) || root.Module != view.roots[0].Module {
			var err error
			tx, err = view.dependencyScope(boundary)
			if err != nil {
				return nil, fmt.Errorf("dependencyScope: %w", err)
			}
		}
		name, err := filepath.Rel(tx.requestedRoot, path)
		if err != nil {
			return nil, fmt.Errorf("Rel: %w", err)
		}
		state, exists := tx.capturedFile(name)
		if !exists {
			state, err = tx.capture(name, view.roots)
			if err != nil {
				return nil, fmt.Errorf("capture: %w", err)
			}
		}
		if !state.exists {
			return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
		}
		if tx == view.observations {
			if proposed, exists := view.proposed[state.path]; exists {
				return bytes.Clone(proposed), nil
			}
		} else if pinned, verified := view.pinned[root.Module]; verified && !bytes.Equal(pinned[filepath.ToSlash(name)], state.data) {
			return nil, fmt.Errorf("source %s differs from verified pinned contents for module %s", path, root.Module)
		}
		return state.data, nil
	}
	return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
}

func (view *tidySourceView) openSourceFile(path string) (io.ReadCloser, error) {
	allowed := view.roots.FileAllowed()
	if allowed != nil && !allowed(path) {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}
	raw, err := view.readSourceFile(path)
	if err != nil {
		return nil, fmt.Errorf("readSourceFile: %w", err)
	}
	return io.NopCloser(bytes.NewReader(raw)), nil
}

func (view *tidySourceView) validateImports(ctx context.Context, files []string, report TidyResult) (map[string]bool, error) {
	owners := make(map[string]bool)
	var queue []v1ImportSource
	seen := make(map[v1ImportSource]bool)
	consumerFiles := make(map[string]bool)
	for _, name := range files {
		path := filepath.Join(view.observations.requestedRoot, name)
		source := v1ImportSource{path: path}
		queue = append(queue, source)
		seen[source] = true
		consumerFiles[path] = true
	}
	allowed := view.roots.FileAllowed()
	var unresolved []v1UnresolvedImport
	for len(queue) > 0 {
		err := ctx.Err()
		if err != nil {
			return nil, fmt.Errorf("Err: %w", err)
		}
		source := queue[0]
		queue = queue[1:]
		var raw []byte
		if source.builtin {
			raw, err = wellknownimports.Content.ReadFile(source.path)
		} else {
			raw, err = view.readSourceFile(source.path)
		}
		if err != nil {
			return nil, fmt.Errorf("readSourceFile: %w", err)
		}
		imports, err := ParseProtoImports(source.path, raw)
		if err != nil {
			return nil, fmt.Errorf("ParseProtoImports: %w", err)
		}
		for _, name := range imports {
			if !ValidProtoImportPath(name) {
				return nil, fmt.Errorf("%s: invalid import %q", source.path, name)
			}
			imported, err := resolveV1ImportSource(name, view.roots, allowed)
			if errors.Is(err, os.ErrNotExist) {
				unresolved = append(unresolved, v1UnresolvedImport{owner: source.path, path: name})
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("resolveV1ImportSource: %s imports %q: %w", source.path, name, err)
			}
			if consumerFiles[source.path] && !imported.builtin {
				for _, root := range view.roots {
					if sourcePathWithin(imported.path, root.Path) && root.allows(imported.path, root.Path) {
						owners[root.Module] = true
						break
					}
				}
			}
			if !seen[imported] {
				seen[imported] = true
				queue = append(queue, imported)
			}
		}
	}
	if len(unresolved) > 0 {
		return nil, fmt.Errorf("cannot resolve imports %v", unresolved)
	}
	err := view.compileChanges(ctx, files, report)
	if err != nil {
		return nil, fmt.Errorf("compileChanges: %w", err)
	}
	return owners, nil
}

func (view *tidySourceView) compileChanges(ctx context.Context, files []string, report TidyResult) error {
	compiled := make(map[string]bool)
	for _, change := range report.Imports {
		if compiled[change.File] {
			continue
		}
		compiled[change.File] = true
		name := ""
		for _, file := range files {
			target, captured := view.observations.capturedTarget(file)
			if !captured {
				return fmt.Errorf("consumer source %q was not captured during planning", file)
			}
			if target != change.File {
				continue
			}
			path := filepath.Join(view.observations.requestedRoot, file)
			for _, root := range view.roots {
				if root.Module == view.roots[0].Module && sourcePathWithin(path, root.Path) {
					relative, err := filepath.Rel(root.Path, path)
					if err != nil {
						return fmt.Errorf("Rel: %w", err)
					}
					name = filepath.ToSlash(relative)
					break
				}
			}
			if name != "" {
				break
			}
		}
		if name == "" {
			return fmt.Errorf("changed source %q has no owning import root", change.File)
		}
		// Independent consumer targets (including owned build copies) are
		// compiled separately; only each target's reachable closure is linked.
		compiler := protocompile.Compiler{Resolver: compilerimports.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: view.roots.Paths(), Accessor: view.openSourceFile,
		})}
		_, err := compiler.Compile(ctx, name)
		if err != nil {
			return fmt.Errorf("Compile: %s: %w", change.File, err)
		}
	}
	return nil
}
