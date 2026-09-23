package api

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestBreakingRejectsLegacyConfig(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "easyp.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("breaking_check:\n  against_git_ref: main\n"), 0o644))
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String(flags.Config.Name, configPath, "")
	ctx := cli.NewContext(&cli.App{Metadata: map[string]any{}}, set, nil)
	ctx.Context = t.Context()
	err := (BreakingCheck{}).action(ctx, logger.NewNop())
	require.ErrorContains(t, err, "ParsePolicy")
}
