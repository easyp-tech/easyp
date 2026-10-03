package api

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestGenerateRejectsConflictingDescriptorOutputFlags(t *testing.T) {
	t.Parallel()
	app := &cli.App{Commands: []*cli.Command{(Generate{}).Command()}}
	err := app.RunContext(t.Context(), []string{"easyp", "generate", "--descriptor_set_out", "all.pb", "--descriptor_set_out_dir", "sets"})
	require.ErrorContains(t, err, "mutually exclusive")
}
