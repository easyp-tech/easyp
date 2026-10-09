package api

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"golang.org/x/mod/sumdb/dirhash"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	gitadapter "github.com/easyp-tech/easyp/internal/adapters/go_git"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestBreakingV1AgainstLockedLegacyBaseline(t *testing.T) {
	// EASYPPATH is process-wide; run independent CLI scenarios sequentially.
	for _, tc := range []struct {
		name             string
		frozen, manifest bool
	}{
		{name: "legacy YAML dependencies"},
		{name: "legacy manifest", manifest: true},
		{name: "frozen legacy YAML", frozen: true},
		{name: "frozen legacy manifest", frozen: true, manifest: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, dependency := t.TempDir(), t.TempDir()
			t.Setenv("EASYPPATH", t.TempDir())
			depProto := "syntax = \"proto3\"; package validate; message Rules { string expression = 1; }\n"
			writeV1GenerateFixture(t, dependency, "validate/validate.proto", depProto)
			runTestGit(t, dependency, "init", "-q")
			runTestGit(t, dependency, "add", ".")
			runTestGit(t, dependency, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "historical dependency")
			commit := strings.TrimSpace(runTestGit(t, dependency, "rev-parse", "HEAD"))
			hash, err := dirhash.Hash1([]string{"validate/validate.proto"}, func(name string) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader(depProto)), nil
			})
			require.NoError(t, err)
			legacyPolicy := "deps: [" + dependency + "@" + commit + "]\ngenerate:\n  inputs: [{directory: {path: proto, root: proto}}]\n  plugins: [{remote: registry/unpinned, out: gen}]\n"
			writeV1GenerateFixture(t, root, "easyp.yaml", legacyPolicy)
			writeV1GenerateFixture(t, root, "easyp.lock", dependency+" "+commit+" "+hash+"\n")
			if tc.manifest {
				writeV1GenerateFixture(t, root, "protobuf.mod", "direct (\n"+dependency+"@"+commit+"\n)\n")
			}
			writeV1GenerateFixture(t, root, "proto/item.proto", "syntax = \"proto3\"; package example.v1; import \"validate/validate.proto\"; message Item { validate.Rules rule = 1; }\n")
			runTestGit(t, root, "init", "-q")
			runTestGit(t, root, "add", ".")
			runTestGit(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "legacy baseline")
			runTestGit(t, root, "branch", "baseline")
			// Neither dependency HEAD nor the current project's graph can compile the
			// old baseline. Only its own recorded revision contains validate.proto.
			require.NoError(t, os.Remove(filepath.Join(dependency, "validate/validate.proto")))
			runTestGit(t, dependency, "add", ".")
			runTestGit(t, dependency, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "remove historical file")
			policy := "version: v1\nbreaking:\n  baseline: git:baseline\n  categories: [FILE]\n"
			manifest := "module example.com/current\nroots proto\n"
			writeV1GenerateFixture(t, root, "easyp.yaml", policy)
			writeV1GenerateFixture(t, root, "protobuf.mod", manifest)
			writeV1GenerateFixture(t, root, "protobuf.lock", "version: 1\nmodules: []\n")
			writeV1GenerateFixture(t, root, "easyp.lock", "not the baseline lock\n")
			writeV1GenerateFixture(t, root, "proto/item.proto", "syntax = \"proto3\"; package example.v1; message Item {}\n")
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			set.String(flagLintDirectoryPath.Name, ".", "")
			set.Bool("frozen", tc.frozen, "")
			ctx := cli.NewContext(&cli.App{}, set, nil)
			ctx.Context = t.Context()
			issues, err := (BreakingCheck{}).checkV1Policies(ctx, logger.NewNop(), filepath.Join(root, "easyp.yaml"), root, root)
			require.NoError(t, err)
			found := false
			for _, issue := range issues {
				found = found || strings.Contains(issue.Message, "was deleted")
			}
			require.True(t, found, "breaking checks must still detect the deleted field: %+v", issues)
			for name, want := range map[string]string{"easyp.yaml": policy, "protobuf.mod": manifest, "easyp.lock": "not the baseline lock\n", "protobuf.lock": "version: 1\nmodules: []\n"} {
				raw, err := os.ReadFile(filepath.Join(root, name))
				require.NoError(t, err)
				require.Equal(t, want, string(raw), "caller file changed: %s", name)
			}
		})
	}
}

func TestBreakingLegacyBaselineRequiresHistoricalLock(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, lock string
		present    bool
	}{
		{name: "missing lock"},
		{name: "empty lock", present: true},
		{name: "invalid lock", lock: "invalid\n", present: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "easyp.yaml", "deps: [example.com/must-not-resolve-head]\n")
			writeV1GenerateFixture(t, root, "item.proto", "syntax = \"proto3\"; message Item {}\n")
			if tc.present {
				writeV1GenerateFixture(t, root, "easyp.lock", tc.lock)
			}
			snapshot := &gitadapter.Snapshot{Root: root, RepositoryRoot: t.TempDir()}
			_, err := breakingImportRoots(t.Context(), gitmodules.New(t.TempDir()), root, ".", []string{"item.proto"}, snapshot)
			require.ErrorContains(t, err, "easyp.lock")
		})
	}
}
