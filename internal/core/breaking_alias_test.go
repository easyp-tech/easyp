package core

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	diskfs "github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

type aliasBreakingWalker struct {
	DirWalker
	root  string
	paths []string
}

func TestBreakingUsesEachRevisionsSourceOpener(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name       string
		categories []string
	}{
		{name: "parser"},
		{name: "descriptors", categories: []string{"FILE"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			current, baseline := t.TempDir(), t.TempDir()
			currentDependency, baselineDependency := t.TempDir(), t.TempDir()
			for _, root := range []string{current, baseline} {
				require.NoError(t, os.WriteFile(filepath.Join(root, "app.proto"), []byte("syntax = \"proto3\"; package app.v1; import \"dep.proto\"; message App { dep.v1.Item item = 1; }"), 0o644))
			}
			require.NoError(t, os.WriteFile(filepath.Join(currentDependency, "dep.proto"), []byte("syntax = \"proto3\"; package dep.v1; message Item { int32 value = 1; }"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(baselineDependency, "dep.proto"), []byte("syntax = \"proto3\"; package dep.v1; message Item { string value = 1; }"), 0o644))
			opener := func(roots []string) func(string) (io.ReadCloser, error) {
				return func(filename string) (io.ReadCloser, error) {
					for _, root := range roots {
						relative, err := filepath.Rel(root, filename)
						if err == nil && filepath.IsLocal(relative) {
							return sourceview.OpenLocal(t.Context(), root, relative)
						}
					}
					return nil, &os.PathError{Op: "open", Path: filename, Err: os.ErrNotExist}
				}
			}
			app := New(Options{Logger: logger.NewNop(), ImportRoots: []string{current, currentDependency}, OpenSourceFile: opener([]string{current, currentDependency}), BreakingCheckConfig: BreakingCheckConfig{Categories: tt.categories}})

			issues, err := app.CompareBreakingWithImports(t.Context(), diskfs.NewFSWalker(current, "."), diskfs.NewFSWalker(baseline, "."), BreakingImports{Paths: []string{baseline, baselineDependency}, Open: opener([]string{baseline, baselineDependency})})

			require.NoError(t, err)
			assert.NotEmpty(t, issues, "changed imported contract must use the historical source opener")
		})
	}
}

func (w aliasBreakingWalker) RootPath() string { return w.root }

func (w aliasBreakingWalker) WalkDir(visit func(string, error) error) error {
	for _, path := range w.paths {
		if err := visit(path, nil); err != nil {
			return err
		}
	}
	return nil
}

func TestBreakingAliasesKeepLogicalFileIdentity(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		paths   []string
		message bool
	}{
		{name: "different import names", paths: []string{"real.proto", "alias.proto"}},
		{name: "repeated same import", paths: []string{"alias.proto", "alias.proto"}},
		{name: "duplicate declarations target first", paths: []string{"real.proto", "alias.proto"}, message: true},
		{name: "duplicate declarations alias first", paths: []string{"alias.proto", "real.proto"}, message: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			source := "syntax = \"proto3\"; package alias.v1;\n"
			if tt.message {
				source += "message Item {}\n"
			}
			require.NoError(t, os.WriteFile(filepath.Join(root, "real.proto"), []byte(source), 0o644))
			require.NoError(t, os.Symlink("real.proto", filepath.Join(root, "alias.proto")))
			app := New(Options{Logger: logger.NewNop(), ImportRoots: []string{root}})

			graph, err := app.compileBreakingGraph(t.Context(), aliasBreakingWalker{DirWalker: diskfs.NewFSWalker(root, "."), root: root, paths: tt.paths})

			if tt.message {
				require.ErrorContains(t, err, "already defined")
				assert.Nil(t, graph)
				return
			}
			require.NoError(t, err)
			assert.Contains(t, graph.targets, "alias.proto")
			if tt.name == "different import names" {
				assert.Len(t, graph.targets, 2)
				assert.Contains(t, graph.targets, "real.proto")
			} else {
				assert.Len(t, graph.targets, 1)
			}
		})
	}
}
