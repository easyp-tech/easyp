package api

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestListV1FilesIncludesReachableWellKnownImport(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	protoFile := filepath.Join(root, "event.proto")
	content := "syntax = \"proto3\";\npackage example.v1;\nimport \"google/protobuf/timestamp.proto\";\nmessage Event { google.protobuf.Timestamp time = 1; }\n"
	require.NoError(t, os.WriteFile(protoFile, []byte(content), 0o644))
	set := flag.NewFlagSet("ls-files", flag.ContinueOnError)
	set.Bool(flagLsFilesIncludeImports.Name, true, "")
	ctx := cli.NewContext(&cli.App{}, set, nil)
	ctx.Context = t.Context()

	listed, err := listV1Files(ctx, root, v1.Module{Name: "example.com/root", Roots: []string{"."}})
	require.NoError(t, err)
	require.Empty(t, listed.Errors)
	require.Len(t, listed.Files, 2)
	require.Equal(t, "event.proto", listed.Files[0].ImportPath)
	require.Equal(t, "google/protobuf/timestamp.proto", listed.Files[1].ImportPath)
	require.Equal(t, "wellknown", listed.Files[1].Source)
}

func TestListV1FilesLocalOnlyDoesNotRequireLock(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "event.proto"), []byte("syntax = \"proto3\";"), 0o644))
	set := flag.NewFlagSet("ls-files", flag.ContinueOnError)
	set.Bool(flagLsFilesIncludeImports.Name, false, "")
	ctx := cli.NewContext(&cli.App{}, set, nil)
	ctx.Context = t.Context()
	module := v1.Module{
		Name:     "example.com/root",
		Roots:    []string{"."},
		Requires: []v1.Requirement{{Module: "example.com/missing", Version: "v1.0.0"}},
	}

	listed, err := listV1Files(ctx, root, module)
	require.NoError(t, err)
	require.Len(t, listed.Files, 1)
	require.Equal(t, "event.proto", listed.Files[0].ImportPath)
	require.Empty(t, listed.Errors)
}

func TestListV1FilesRejectsDuplicateDependencyImports(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dependency := t.TempDir()
	replacement, err := filepath.Rel(root, dependency)
	require.NoError(t, err)
	writeV1GenerateFixture(t, root, "event.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, dependency, "protobuf.mod", "module example.com/dep\n")
	writeV1GenerateFixture(t, dependency, "event.proto", "syntax = \"proto3\";\n")
	set := flag.NewFlagSet("ls-files", flag.ContinueOnError)
	set.Bool(flagLsFilesIncludeImports.Name, true, "")
	ctx := cli.NewContext(&cli.App{}, set, nil)
	ctx.Context = t.Context()
	module := v1.Module{
		Name: "example.com/root", Roots: []string{"."},
		Requires: []v1.Requirement{{Module: "example.com/dep", Version: "v1.0.0"}},
		Replaces: []v1.Replacement{{Module: "example.com/dep", Target: replacement}},
	}

	_, err = listV1Files(ctx, root, module)
	require.ErrorContains(t, err, "duplicate import path")
	require.ErrorContains(t, err, "event.proto")
}
