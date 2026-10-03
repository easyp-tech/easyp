package api

import (
	"flag"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestBreakingV1FileCategoryReachesChecker(t *testing.T) {
	// buildCore configures process-wide comment-ignore state.
	// Keep CLI scenarios sequential rather than sharing that state in parallel.
	const header = "syntax = \"proto3\";\npackage example.v1;\n"
	const declarations = `message Item {
  string id = 1;
  oneof value {
    string text = 2;
  }
}
enum State {
  STATE_UNSPECIFIED = 0;
}
service ItemService {
  rpc Get(Item) returns (Item);
}
`
	tests := []struct {
		name       string
		categories string
		removeID   bool
		wantMoved  int
		wantDelete int
	}{
		{name: "default ignores moves"},
		{name: "FILE detects moves", categories: "  categories: [FILE]\n", wantMoved: 4},
		{name: "FILE preserves field checks", categories: "  categories: [FILE]\n", removeID: true, wantMoved: 4, wantDelete: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			configPath := filepath.Join(root, "easyp.yaml")
			writeV1GenerateFixture(t, root, "easyp.yaml", "version: v1\nbreaking:\n  baseline: git:baseline\n"+tt.categories)
			writeV1GenerateFixture(t, root, "old.proto", header+declarations)
			runTestGit(t, root, "init", "-q")
			runTestGit(t, root, "add", ".")
			runTestGit(t, root, "-c", "user.name=EasyP test", "-c", "user.email=test@example.test", "commit", "-qm", "baseline")
			runTestGit(t, root, "branch", "baseline")
			current := declarations
			if tt.removeID {
				current = strings.ReplaceAll(current, "  string id = 1;\n", "")
			}
			writeV1GenerateFixture(t, root, "old.proto", header)
			writeV1GenerateFixture(t, root, "new.proto", header+current)
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			set.String(flagLintDirectoryPath.Name, ".", "")
			ctx := cli.NewContext(&cli.App{}, set, nil)
			ctx.Context = t.Context()

			issues, err := (BreakingCheck{}).checkV1Policies(ctx, logger.NewNop(), configPath, root, root)

			require.NoError(t, err)
			moved, deleted := 0, 0
			for _, issue := range issues {
				if strings.Contains(issue.Message, "was moved") {
					moved++
					assert.Contains(t, issue.Message, "old.proto")
					assert.Contains(t, issue.Message, "new.proto")
				}
				if strings.Contains(issue.Message, "was deleted") {
					deleted++
				}
			}
			assert.Equal(t, tt.wantMoved, moved, "%+v", issues)
			assert.Equal(t, tt.wantDelete, deleted, "%+v", issues)
			assert.Len(t, issues, tt.wantMoved+tt.wantDelete)
		})
	}
}
