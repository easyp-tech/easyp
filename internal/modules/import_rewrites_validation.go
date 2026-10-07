package modules

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bufbuild/protocompile"
	compilerimports "github.com/bufbuild/protocompile/wellknownimports"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/wellknownimports"
)

// tidySourceView reads the proposed consumer bytes and captures unchanged
// reachable inputs for the same pre-commit state check as the project writer.
type tidySourceView struct {
	tx       *resolvedFilesTransaction
	roots    SourceRoots
	proposed map[string][]byte
	scopes   map[string]*resolvedFilesTransaction
	pinned   map[string]map[string][]byte
	old      []tidyOldNamespace
	mu       sync.Mutex
}

func newTidySourceView(tx *resolvedFilesTransaction, roots SourceRoots) *tidySourceView {
	return &tidySourceView{tx: tx, roots: roots, proposed: make(map[string][]byte), scopes: make(map[string]*resolvedFilesTransaction), pinned: make(map[string]map[string][]byte)}
}

func (view *tidySourceView) retainPinnedSources(source *importRootSource, lock v1.Lock) {
	for _, entry := range lock.Modules {
		fetched := source.fetched[[2]string{entry.Source, strings.ToLower(entry.Commit)}]
		if fetched.Inspection != nil && !fetched.Inspection.Provisional {
			view.pinned[entry.Source] = make(map[string][]byte)
			for _, file := range fetched.Inspection.Files {
				view.pinned[entry.Source][file.Path] = file.Content
			}
		}
	}
}

type tidyCachedModule struct {
	directory string
	module    v1.Module
}

func (view *tidySourceView) cachedSources(ctx context.Context, lock v1.Lock, repository Repository) (SourceRoots, error) {
	var cached []tidyCachedModule
	for _, entry := range lock.Modules {
		directory, module, err := repository.Cached(entry)
		if err != nil {
			return nil, fmt.Errorf("Cached: %w", err)
		}
		directory, err = filepath.Abs(directory)
		if err != nil {
			return nil, fmt.Errorf("Abs: %w", err)
		}
		// Discover the directory, then capture and verify its metadata before
		// using the returned requirements or roots to build the source graph.
		err = view.observeModuleMetadata(directory, module)
		if err != nil {
			return nil, fmt.Errorf("observeModuleMetadata: %w", err)
		}
		cached = append(cached, tidyCachedModule{directory: directory, module: module})
	}
	err := view.tx.verifyCapturedInputs()
	if err != nil {
		return nil, fmt.Errorf("verifyCapturedInputs: %w", err)
	}
	err = tidyInputVerifier(ctx, lock, repository)()
	if err != nil {
		return nil, fmt.Errorf("tidyInputVerifier: %w", err)
	}
	err = view.tx.verifyCapturedInputs()
	if err != nil {
		return nil, fmt.Errorf("verifyCapturedInputs: %w", err)
	}
	var roots SourceRoots
	for _, dependency := range cached {
		err := ValidateRequirements(dependency.module.Requires, lock)
		if err != nil {
			return nil, fmt.Errorf("ValidateRequirements: %w", err)
		}
		selected, err := ModuleSources(dependency.directory, dependency.module)
		if err != nil {
			return nil, fmt.Errorf("ModuleSources: %w", err)
		}
		roots = append(roots, selected...)
	}
	return roots, nil
}

func (view *tidySourceView) observeModuleMetadata(directory string, module v1.Module) error {
	tx, err := view.dependencyScope(directory)
	if err != nil {
		return fmt.Errorf("dependencyScope: %w", err)
	}
	for _, root := range append([]string{"."}, module.Roots...) {
		// Native module locations and Buf workspace member metadata are
		// ancestors of the selected roots. Capture those validation inputs,
		// including their absence, without reading unrelated proto bodies.
		for directory := root; ; directory = filepath.Dir(directory) {
			for _, name := range []string{v1.ModuleFile, "easyp.yaml", "buf.yaml", "buf.work.yaml", "buf.lock"} {
				_, err := tx.capture(filepath.Join(directory, name), nil)
				if err != nil {
					return fmt.Errorf("capture: %w", err)
				}
			}
			if directory == "." {
				break
			}
		}
	}
	return nil
}

func (view *tidySourceView) dependencyScope(boundary string) (*resolvedFilesTransaction, error) {
	tx, exists := view.scopes[boundary]
	if exists {
		return tx, nil
	}
	tx, err := newResolvedFilesTransaction(boundary)
	if err != nil {
		return nil, fmt.Errorf("newResolvedFilesTransaction: %w", err)
	}
	view.scopes[boundary] = tx
	view.tx.inputs = append(view.tx.inputs, tx)
	return tx, nil
}

func (view *tidySourceView) readSourceFile(path string) ([]byte, error) {
	view.mu.Lock()
	defer view.mu.Unlock()
	for _, root := range view.roots {
		if !sourcePathWithin(path, root.Path) || !root.allows(path, root.Path) {
			continue
		}
		tx := view.tx
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
		state, exists := tx.expected[name]
		if !exists {
			state, err = tx.capture(name, view.roots)
			if err != nil {
				return nil, fmt.Errorf("capture: %w", err)
			}
		}
		if !state.exists {
			return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
		}
		if tx == view.tx {
			if proposed, exists := view.proposed[state.resolution.Path]; exists {
				return proposed, nil
			}
		} else if pinned, verified := view.pinned[root.Module]; verified && !bytes.Equal(pinned[filepath.ToSlash(name)], state.data) {
			return nil, fmt.Errorf("source %s differs from verified pinned contents for module %s", path, root.Module)
		}
		return state.data, nil
	}
	return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
}

type tidyCachedVerifier interface {
	VerifyCached(context.Context, v1.Lock) error
}

func tidyInputVerifier(ctx context.Context, lock v1.Lock, repository Repository) func() error {
	return func() error {
		err := ctx.Err()
		if err != nil {
			return fmt.Errorf("Err: %w", err)
		}
		if verifier, supported := repository.(tidyCachedVerifier); supported {
			err = verifier.VerifyCached(ctx, lock)
			if err != nil {
				return fmt.Errorf("VerifyCached: %w", err)
			}
		} else {
			// The base Cache contract verifies installed bytes against the lock.
			// Callers recheck snapshots after this potentially repairing operation.
			err = repository.Install(ctx, lock)
			if err != nil {
				return fmt.Errorf("Install: %w", err)
			}
		}
		return nil
	}
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

func validateTidyImports(ctx context.Context, files []string, report TidyResult, view *tidySourceView) (map[string]bool, error) {
	owners := make(map[string]bool)
	var queue []v1ImportSource
	seen := make(map[v1ImportSource]bool)
	consumerFiles := make(map[string]bool)
	for _, name := range files {
		path := filepath.Join(view.tx.requestedRoot, name)
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
	err := compileTidyChanges(ctx, files, report, view)
	if err != nil {
		return nil, fmt.Errorf("compileTidyChanges: %w", err)
	}
	return owners, nil
}

func compileTidyChanges(ctx context.Context, files []string, report TidyResult, view *tidySourceView) error {
	compiled := make(map[string]bool)
	for _, change := range report.Imports {
		if compiled[change.File] {
			continue
		}
		compiled[change.File] = true
		name := ""
		for _, file := range files {
			if view.tx.expected[file].resolution.Path != change.File {
				continue
			}
			path := filepath.Join(view.tx.requestedRoot, file)
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
