package api

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestListV1Files(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		includeImports bool
		requires       []v1.Requirement
		wantImports    []string
		wantSources    []string
	}{
		{
			name: "reachable well-known import", includeImports: true,
			wantImports: []string{"event.proto", "google/protobuf/timestamp.proto"},
			wantSources: []string{"workspace", "wellknown"},
		},
		{
			name: "local-only ignores missing lock", requires: []v1.Requirement{{Module: "example.com/missing", Version: "v1.0.0"}},
			wantImports: []string{"event.proto"}, wantSources: []string{"workspace"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "event.proto", `syntax = "proto3"; package example.v1; import "google/protobuf/timestamp.proto"; message Event { google.protobuf.Timestamp time = 1; }`)
			module := v1.Module{Name: "example.com/root", Roots: []string{"."}, Requires: tt.requires}

			listed, err := listV1Files(t.Context(), root, module, tt.includeImports, nil)

			require.NoError(t, err)
			assert.Empty(t, listed.Errors)
			var imports, sources []string
			for _, file := range listed.Files {
				imports = append(imports, file.ImportPath)
				sources = append(sources, file.Source)
			}
			assert.Equal(t, tt.wantImports, imports)
			assert.Equal(t, tt.wantSources, sources)
		})
	}
}

func TestListV1FilesRejectsDuplicateDependencyImports(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		importPath string
	}{
		{name: "root_import", importPath: "event.proto"},
		{name: "nested_import", importPath: "events/v1/event.proto"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			dependency := t.TempDir()
			replacement, err := filepath.Rel(root, dependency)
			require.NoError(t, err)
			writeV1GenerateFixture(t, root, tt.importPath, "syntax = \"proto3\";\n")
			writeV1GenerateFixture(t, dependency, "protobuf.mod", "module example.com/dep\n")
			writeV1GenerateFixture(t, dependency, tt.importPath, "syntax = \"proto3\";\n")
			module := v1.Module{
				Name: "example.com/root", Roots: []string{"."},
				Requires: []v1.Requirement{{Module: "example.com/dep", Version: "v1.0.0"}},
				Replaces: []v1.Replacement{{Module: "example.com/dep", Target: replacement}},
			}

			_, err = listV1Files(t.Context(), root, module, true, nil)

			require.ErrorContains(t, err, "duplicate import path")
			require.ErrorContains(t, err, tt.importPath)
		})
	}
}

// The CLI owns process flags, environment, and cwd, so these cases are sequential.
func TestLsFilesLocalOnlyWithoutCacheEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
	}{
		{name: "local module", manifest: "module example.com/root\n"},
		{name: "uninstalled dependency", manifest: "module example.com/root\nrequire example.com/missing v1.0.0\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "protobuf.mod", tt.manifest)
			writeV1GenerateFixture(t, root, "event.proto", "syntax = \"proto3\";\n")
			t.Chdir(root)
			t.Setenv("HOME", "")
			t.Setenv("EASYPPATH", "")
			flags := flag.NewFlagSet("ls-files", flag.ContinueOnError)
			flags.Bool(flagLsFilesIncludeImports.Name, false, "")
			ctx := cli.NewContext(cli.NewApp(), flags, nil)
			ctx.Context = t.Context()

			err := (LsFiles{}).Action(ctx)

			require.NoError(t, err)
		})
	}
}
