package generation

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/workspace"
)

// v1WorkspaceModuleIndex snapshots declarations lazily for one generation run.
// Dependency graphs and selected module metadata remain consumer-owned.
type v1WorkspaceModuleIndex struct {
	root     string
	walk     func(string, func(string, fs.DirEntry, error) error) error
	readFile func(string, string) ([]byte, error)
	once     sync.Once
	modules  map[string]v1WorkspaceModuleDeclaration
	err      error
}

type v1WorkspaceModuleDeclaration struct {
	directory string
	duplicate string
}

func newV1WorkspaceModuleIndex(root string) *v1WorkspaceModuleIndex {
	return &v1WorkspaceModuleIndex{root: root, walk: workspace.Walk, readFile: workspace.ReadFile}
}

func (index *v1WorkspaceModuleIndex) find(name string) (string, error) {
	index.once.Do(index.load)
	declaration := index.modules[name]
	if declaration.duplicate != "" {
		return declaration.directory, fmt.Errorf("module identity %q is declared in both %s and %s", name, declaration.directory, declaration.duplicate)
	}
	return declaration.directory, index.err
}

func (index *v1WorkspaceModuleIndex) load() {
	index.modules = make(map[string]v1WorkspaceModuleDeclaration)
	index.err = index.walk(index.root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Source aliases are validated after selection, independently of declarations.
			if filepath.Base(path) != v1.ModuleFile {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() {
			if workspace.SkipDirectory(index.root, path) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != v1.ModuleFile {
			return nil
		}
		manifest, err := index.readFile(index.root, path)
		if err != nil {
			return fmt.Errorf("ReadFile: %w", err)
		}
		if !v1.IsModuleManifest(manifest) {
			return nil
		}
		module, err := v1.ParseModule(bytes.NewReader(manifest))
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		declaration := index.modules[module.Name]
		if declaration.directory == "" {
			declaration.directory = filepath.Dir(path)
		} else if declaration.duplicate == "" {
			// Keep the first duplicate without failing unrelated identity lookups.
			// It also precedes any later walk/read/parse error, just as in a direct walk.
			declaration.duplicate = filepath.Dir(path)
		}
		index.modules[module.Name] = declaration
		return nil
	})
}
