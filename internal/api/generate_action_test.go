package api

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestGenerateRequiresV1Config(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	require.NoError(t, os.WriteFile(filepath.Join(root, "easyp.yaml"), []byte("generate:\n  inputs: []\n"), 0o644))

	ctx := cli.NewContext(&cli.App{Metadata: map[string]any{}}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
	ctx.Context = t.Context()
	err := (Generate{}).Action(ctx)
	require.ErrorContains(t, err, "easyp.gen.yaml")
}
