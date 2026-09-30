package api

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestBreakingIgnoreUnstableConfigReachesCore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		legacy, ignore, explicit bool
		wantCount                int
	}{
		{name: "v1_enabled", ignore: true, wantCount: 1},
		{name: "v1_disabled", wantCount: 2},
		{name: "legacy_enabled", legacy: true, ignore: true, wantCount: 1},
		{name: "legacy_disabled", legacy: true, wantCount: 2},
		{name: "explicit_baseline_preserves_options", ignore: true, explicit: true, wantCount: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.BreakingCheck{IgnoreUnstable: tt.ignore, AgainstGitRef: "release", Use: []string{core.BreakingCheckFilesCheck}}
			if !tt.legacy {
				policy, err := v1.ParsePolicy(strings.NewReader(fmt.Sprintf("version: v1\nbreaking:\n  baseline: git:release\n  categories: [FILE]\n  ignore_unstable: %t\n", tt.ignore)))
				require.NoError(t, err)
				set := flag.NewFlagSet("test", flag.ContinueOnError)
				set.String(flagAgainstBranchName.Name, "master", "")
				if tt.explicit {
					require.NoError(t, set.Set(flagAgainstBranchName.Name, "override"))
				}
				cfg, err = resolveV1BreakingConfig(cli.NewContext(&cli.App{}, set, nil), policy)
				require.NoError(t, err)
				assert.Equal(t, tt.ignore, cfg.IgnoreUnstable)
				if tt.explicit {
					assert.Equal(t, "override", cfg.AgainstGitRef)
				} else {
					assert.Equal(t, "release", cfg.AgainstGitRef)
				}
			}
			baseline, current := t.TempDir(), t.TempDir()
			for _, pkg := range []string{"stable.v1", "unstable.v1beta1"} {
				content := "syntax = \"proto3\"; package " + pkg + "; message Item {}"
				writeV1GenerateFixture(t, baseline, pkg+"_old.proto", content)
				writeV1GenerateFixture(t, current, pkg+"_new.proto", content)
			}
			app, err := buildCore(logger.NewNop(), config.Config{BreakingCheck: cfg}, nil)
			require.NoError(t, err)
			issues, err := app.CompareBreaking(t.Context(), fs.NewFSWalker(current, "."), fs.NewFSWalker(baseline, "."), nil)
			require.NoError(t, err)
			assert.Len(t, issues, tt.wantCount)
			for _, issue := range issues {
				assert.Contains(t, issue.Message, "was moved")
			}
		})
	}
}

func TestBreakingV1IgnoreUnstableModuleSnapshot(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		ignore    bool
		wantCount int
	}{
		{name: "enabled", ignore: true, wantCount: 2},
		{name: "disabled", wantCount: 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "easyp.yaml", fmt.Sprintf("version: v1\nbreaking:\n  baseline: git:baseline\n  categories: [FILE]\n  ignore_unstable: %t\n", tt.ignore))
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/test\nroots proto\n")
			for _, pkg := range []string{"stable.v1", "unstable.v1beta1"} {
				writeV1GenerateFixture(t, root, "proto/"+pkg+"_old.proto", "syntax = \"proto3\"; package "+pkg+"; message Item { string id = 1; }")
			}
			runTestGit(t, root, "init", "-q")
			runTestGit(t, root, "add", ".")
			runTestGit(t, root, "-c", "user.name=EasyP test", "-c", "user.email=test@example.test", "commit", "-qm", "baseline fixture")
			runTestGit(t, root, "branch", "baseline")
			for _, pkg := range []string{"stable.v1", "unstable.v1beta1"} {
				writeV1GenerateFixture(t, root, "proto/"+pkg+"_old.proto", "syntax = \"proto3\"; package "+pkg+";")
				writeV1GenerateFixture(t, root, "proto/"+pkg+"_new.proto", "syntax = \"proto3\"; package "+pkg+"; message Item {}")
			}
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			set.String(flagLintDirectoryPath.Name, ".", "")
			ctx := cli.NewContext(&cli.App{}, set, nil)
			ctx.Context = t.Context()
			issues, err := (BreakingCheck{}).checkV1Policies(ctx, logger.NewNop(), filepath.Join(root, "easyp.yaml"), root, filepath.Join(root, "proto"))
			require.NoError(t, err)
			assert.Len(t, issues, tt.wantCount, "%+v", issues)
			if tt.ignore {
				for _, issue := range issues {
					assert.Equal(t, "proto/stable.v1_old.proto", issue.Path)
				}
			}
		})
	}
}
