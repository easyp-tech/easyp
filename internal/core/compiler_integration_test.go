package core

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protodesc"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestPrepareGenerationPreservesVirtualSourceInventory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sources := map[string]string{
		"shared.proto": `syntax = "proto3";
package shared;
message Common {}
`,
		"api/first.proto": `syntax = "proto3";
package api.v1;
import "shared.proto";

// First is a documented target.
message First {
  shared.Common value = 1;
}
`,
		"api/second.proto": `syntax = "proto3";
package api.v1;
import "shared.proto";
message Second { shared.Common value = 1; }
`,
	}
	app := New(Options{
		Logger: logger.NewNop(),
		Inputs: Inputs{InputFilesDir: []InputFilesDir{{Root: ".", Path: "api"}}},
		WalkSourceFiles: func(_ string, visit func(string) error) (bool, error) {
			for _, name := range []string{"api/second.proto", "api/first.proto"} {
				if err := visit(filepath.Join(root, filepath.FromSlash(name))); err != nil {
					return true, err
				}
			}
			return true, nil
		},
		OpenSourceFile: func(path string) (io.ReadCloser, error) {
			name, err := filepath.Rel(root, path)
			if err != nil {
				return nil, err
			}
			source, ok := sources[filepath.ToSlash(name)]
			if !ok {
				return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
			}
			return io.NopCloser(strings.NewReader(source)), nil
		},
	})

	plan, err := app.PrepareGeneration(t.Context(), root)

	require.NoError(t, err)
	set := plan.DescriptorSet(true)
	assert.Equal(t, []string{"shared.proto", "api/second.proto", "api/first.proto"}, descriptorNames(set.File))
	assert.Equal(t, []string{"api/second.proto", "api/first.proto"}, descriptorNames(plan.DescriptorSet(false).File))
	registry, err := protodesc.NewFiles(set)
	require.NoError(t, err)
	file, err := registry.FindFileByPath("api/first.proto")
	require.NoError(t, err)
	location := file.SourceLocations().ByDescriptor(file.Messages().ByName("First"))
	assert.Equal(t, 5, location.StartLine)
	assert.Equal(t, 0, location.StartColumn)
	assert.Contains(t, location.LeadingComments, "First is a documented target.")
	_, err = os.Stat(filepath.Join(root, "api/first.proto"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestPrepareGenerationPreservesResolverDenial(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeGenerationSources(t, root, map[string]string{
		"api/item.proto": `syntax = "proto3"; import "dep.proto"; message Item { Dep dep = 1; }`,
		"dep.proto":      `syntax = "proto3"; message Dep {}`,
	})
	app := testCoreWithPlugins(nil, &captureExecutor{})
	app.sourceFileOpen = func(path string) (io.ReadCloser, error) {
		if path == filepath.Join(root, "dep.proto") {
			return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrPermission}
		}
		return os.Open(path)
	}

	plan, err := app.PrepareGeneration(t.Context(), root)

	require.ErrorIs(t, err, os.ErrPermission)
	assert.Nil(t, plan)
}

func TestBreakingDescriptorsUseBaselineSourceLocations(t *testing.T) {
	t.Parallel()

	before := writeProfileProto(t, "baseline", "syntax = \"proto3\";\npackage api.v1;\nmessage Item {\n  string removed = 1;\n}\n")
	after := writeProfileProto(t, "current", "syntax = \"proto3\";\npackage api.v1;\nmessage Item {}\n")
	app := New(Options{Logger: logger.NewNop()})
	oldGraph, err := app.compileBreakingGraph(t.Context(), aliasBreakingWalker{root: before, paths: []string{"schema.proto"}})
	require.NoError(t, err)
	newGraph, err := app.compileBreakingGraph(t.Context(), aliasBreakingWalker{root: after, paths: []string{"schema.proto"}})
	require.NoError(t, err)

	issues := compareBreakingGraphs(oldGraph, newGraph, breakingProfiles{file: true, source: true, wireJSON: true, wire: true})

	require.NotEmpty(t, issues)
	for _, issue := range issues {
		if issue.RuleName == "FIELD_NO_DELETE" {
			assert.Equal(t, "schema.proto", issue.Path)
			assert.Equal(t, 4, issue.Position.Line)
			assert.Equal(t, 3, issue.Position.Column)
			return
		}
	}
	t.Fatalf("FIELD_NO_DELETE issue missing: %+v", issues)
}
