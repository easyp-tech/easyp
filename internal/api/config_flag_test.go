package api

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestPolicyCommandsRejectInvalidConfigSelection(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "lint empty", args: []string{"lint", "--config", ""}, want: "--config must not be empty"},
		{name: "lint generation config", args: []string{"lint", "--config", "easyp.gen.yaml"}, want: "easyp.gen.yaml is a generation config; this command expects a policy config"},
		{name: "lint generation profile", args: []string{"lint", "--config", "private.easyp.gen.yaml"}, want: "private.easyp.gen.yaml is a generation config"},
		{name: "breaking empty", args: []string{"breaking", "--config", ""}, want: "--config must not be empty"},
		{name: "breaking lock", args: []string{"breaking", "--config", "protobuf.lock"}, want: "protobuf.lock is a module lock; this command expects a policy config"},
		{name: "validate empty", args: []string{"validate-config", "--config", ""}, want: "--config must not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configFlagApp().RunContext(t.Context(), append([]string{"easyp"}, tt.args...))
			require.ErrorContains(t, err, tt.want)
		})
	}
}

func TestValidateConfigSelectsNamedGenerationProfile(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	path := filepath.Join(root, "private.easyp.gen.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: v1\nplugins:\n  - name: go\n    out: gen\n"), 0o644))

	err := configFlagApp().RunContext(t.Context(), []string{"easyp", "validate-config", "--config", "private.easyp.gen.yaml"})

	require.NoError(t, err)
}

func configFlagApp() *cli.App {
	commands := []*cli.Command{(Lint{}).Command(), (BreakingCheck{}).Command(), (Generate{}).Command(), (Validate{}).Command()}
	for _, command := range commands {
		// urfave's automatic HelpFlag is global mutable state across apps.
		command.HideHelp = true
	}
	return &cli.App{Commands: commands, HideHelp: true, Writer: io.Discard, ErrWriter: io.Discard}
}
