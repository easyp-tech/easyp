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

// The baseline collision check runs while both revisions compile, but a
// collision still fails the run instead of reporting the breaking changes.
func TestBreakingV1RejectsBaselineImportCollision(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "easyp.yaml")
	writeV1GenerateFixture(t, root, "easyp.yaml", "version: v1\nbreaking:\n  baseline: git:baseline\n  categories: [FILE]\n")
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/root\nroots proto\nrequire example.com/dep v1.0.0\nreplace example.com/dep => ./deps/dep\n")
	writeV1GenerateFixture(t, root, "proto/root/v1/root.proto", "syntax = \"proto3\"; package root.v1; import \"dep/v1/dep.proto\"; message Root { dep.v1.Dep value = 1; string name = 2; }\n")
	writeV1GenerateFixture(t, root, "proto/dep/v1/dep.proto", "syntax = \"proto3\"; package dep.v1; message Dep {}\n")
	writeV1GenerateFixture(t, root, "deps/dep/protobuf.mod", "module example.com/dep\nroots src\n")
	writeV1GenerateFixture(t, root, "deps/dep/src/dep/v1/dep.proto", "syntax = \"proto3\"; package dep.v1; message Dep {}\n")
	runTestGit(t, root, "init", "-q")
	runTestGit(t, root, "add", ".")
	runTestGit(t, root, "-c", "user.name=EasyP test", "-c", "user.email=test@example.test", "commit", "-qm", "baseline")
	runTestGit(t, root, "branch", "baseline")
	// The current revision drops the duplicate and a field, so only the
	// baseline collides and the comparison alone would report a finding.
	require.NoError(t, os.Remove(filepath.Join(root, "proto/dep/v1/dep.proto")))
	writeV1GenerateFixture(t, root, "proto/root/v1/root.proto", "syntax = \"proto3\"; package root.v1; import \"dep/v1/dep.proto\"; message Root { dep.v1.Dep value = 1; }\n")
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String(flagLintDirectoryPath.Name, ".", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)
	ctx.Context = t.Context()

	issues, err := (BreakingCheck{}).checkV1Policies(ctx, logger.NewNop(), configPath, root, root)

	require.ErrorContains(t, err, "baseline baseline module .: CheckSourceCollisions: duplicate import path \"dep/v1/dep.proto\"")
	require.Empty(t, issues)
}
