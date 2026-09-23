package api

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestLintV1RejectsLegacyConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	configPath := filepath.Join(root, "easyp.yaml")
	err := os.WriteFile(configPath, []byte("lint:\n  use: [FILE_LOWER_SNAKE_CASE]\n"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(root, "file.proto"), []byte("syntax = \"proto3\";\npackage example;\n"), 0o644)
	require.NoError(t, err)
	ctx := cli.NewContext(&cli.App{}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
	ctx.Context = t.Context()

	err = (Lint{}).actionV1(ctx, logger.NewNop(), configPath, root, root)
	require.Error(t, err)
}

func TestLintV1UsesSelectedConfigFile(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		header string
	}{
		{name: "explicit version", header: "version: v1\n"},
		{name: "default version"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			configPath := filepath.Join(root, "strict.yaml")
			err := os.WriteFile(configPath, []byte(tc.header+"issues:\n  exclude-rules:\n    - {}\n"), 0o644)
			require.NoError(t, err)
			// An unrelated default config must not override the explicitly selected one.
			err = os.WriteFile(filepath.Join(root, "easyp.yaml"), []byte("lint:\n  use: [FILE_LOWER_SNAKE_CASE]\n"), 0o644)
			require.NoError(t, err)
			err = os.WriteFile(filepath.Join(root, "file.proto"), []byte("syntax = \"proto3\";\npackage example;\n"), 0o644)
			require.NoError(t, err)
			ctx := cli.NewContext(&cli.App{}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
			ctx.Context = t.Context()

			err = (Lint{}).actionV1(ctx, logger.NewNop(), configPath, root, root)
			require.NoError(t, err)
		})
	}
}

func TestLintV1ResolvesLocalDependencyImport(t *testing.T) {
	root := t.TempDir()
	dep := t.TempDir()
	replacement, err := filepath.Rel(root, dep)
	require.NoError(t, err)
	writeV1GenerateFixture(t, root, "easyp.yaml", "version: v1\nlinters:\n  default: MINIMAL\n")
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/root\nroots proto\nrequire example.com/dep v1.0.0\nreplace example.com/dep => "+replacement+"\n")
	writeV1GenerateFixture(t, root, "proto/root/v1/root.proto", "syntax = \"proto3\"; package root.v1; import \"dep/v1/dep.proto\"; message Root { dep.v1.Dep value = 1; }\n")
	writeV1GenerateFixture(t, dep, "protobuf.mod", "module example.com/dep\nroots src\n")
	writeV1GenerateFixture(t, dep, "src/dep/v1/dep.proto", "syntax = \"proto3\"; package dep.v1; message Dep {}\n")
	ctx := cli.NewContext(&cli.App{}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
	ctx.Context = t.Context()

	err = (Lint{}).actionV1(ctx, logger.NewNop(), filepath.Join(root, "easyp.yaml"), root, filepath.Join(root, "proto"))
	require.NoError(t, err)
}
