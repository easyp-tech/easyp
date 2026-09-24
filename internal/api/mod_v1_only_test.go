package api

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestModCommandsRejectLegacyManifest(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(Mod, *cli.Context) error
	}{
		{name: "download", run: Mod.Download},
		{name: "update", run: Mod.Update},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			require.NoError(t, os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte("direct (\n)\n"), 0o644))
			ctx := cli.NewContext(&cli.App{Metadata: map[string]any{}}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
			ctx.Context = t.Context()
			err := tc.run(Mod{}, ctx)
			require.ErrorContains(t, err, "ParseModule")
		})
	}
}
