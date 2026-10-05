package api

import (
	"flag"
	"github.com/easyp-tech/easyp/internal/logger"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestFrozenMutationCommandsRefuseBeforeAccess(t *testing.T) {
	// Command constructors still contain shared legacy flags; keep parsers sequential.
	for _, args := range [][]string{
		{"get", "--frozen", "example.com/missing@v1.0.0"},
		{"mod", "--frozen", "update"},
		{"mod", "update", "--frozen"},
		{"mod", "tidy", "--frozen"},
		{"init", "--frozen", "--dir", "/must-not-be-created"},
		{"migrate", "--frozen", "--write"},
	} {
		t.Run(args[0]+"/"+args[1], func(t *testing.T) {
			app := &cli.App{Writer: io.Discard, ErrWriter: io.Discard, Commands: []*cli.Command{
				(Get{}).Command(), (Mod{}).Command(), (Init{}).Command(), (Migrate{}).Command(),
			}}
			err := app.RunContext(t.Context(), append([]string{"easyp"}, args...))
			require.ErrorContains(t, err, "not allowed in frozen mode")
		})
	}
}

func TestFrozenBreakingValidatesEmptyExplicitModule(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/empty\n")
	_, err := breakingImportRootsMode(t.Context(), nil, root, ".", nil, nil, true)
	require.ErrorContains(t, err, "protobuf.lock")
	writeV1GenerateFixture(t, root, "protobuf.lock", "version: 1\nmodules: []\n")
	_, err = breakingImportRootsMode(t.Context(), nil, root, ".", nil, nil, true)
	require.NoError(t, err)
}

func TestFrozenLintValidatesEmptyDescendantModule(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/root\n")
	writeV1GenerateFixture(t, root, "protobuf.lock", "version: 1\nmodules: []\n")
	writeV1GenerateFixture(t, root, "easyp.yaml", "version: v1\nlinters:\n  default: MINIMAL\n")
	writeV1GenerateFixture(t, root, "child/protobuf.mod", "module example.com/child\n")
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.Bool("frozen", true, "")
	set.String("path", ".", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)
	ctx.Context = t.Context()
	err := (Lint{}).actionV1(ctx, logger.NewNop(), filepath.Join(root, "easyp.yaml"), root, root)
	require.ErrorContains(t, err, "protobuf.lock")
	require.ErrorContains(t, err, "child")
}

func TestFrozenLintExcludedSourcesStillRequireManifest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "easyp.yaml", "version: v1\nissues:\n  exclude-rules: [{}]\n")
	writeV1GenerateFixture(t, root, "api.proto", "syntax = \"proto3\";\n")
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.Bool("frozen", true, "")
	set.String("path", ".", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)
	ctx.Context = t.Context()
	err := (Lint{}).actionV1(ctx, logger.NewNop(), filepath.Join(root, "easyp.yaml"), root, root)
	require.ErrorContains(t, err, "protobuf.mod")
	require.ErrorContains(t, err, "frozen")
}
