package flags

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestFrozenPreservesTrueAcrossLineage(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		args []string
		want bool
	}{
		{name: "default", args: []string{"mod", "download"}},
		{name: "global", args: []string{"--frozen", "mod", "download"}, want: true},
		{name: "parent", args: []string{"mod", "--frozen", "download"}, want: true},
		{name: "leaf", args: []string{"mod", "download", "--frozen"}, want: true},
		{name: "global true local false", args: []string{"--frozen", "mod", "--frozen=false", "download", "--frozen=false"}, want: true},
		{name: "parent true leaf false", args: []string{"mod", "--frozen", "download", "--frozen=false"}, want: true},
		{name: "all false", args: []string{"--frozen=false", "mod", "download", "--frozen=false"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			called := false
			app := &cli.App{HideHelp: true, HideVersion: true, Flags: []cli.Flag{Frozen()}, Commands: []*cli.Command{{
				Name: "mod", HideHelp: true, Flags: []cli.Flag{Frozen()}, Subcommands: []*cli.Command{{
					Name: "download", HideHelp: true, Flags: []cli.Flag{Frozen()}, Action: func(ctx *cli.Context) error {
						called = true
						require.Equal(t, tt.want, IsFrozen(ctx))
						return nil
					},
				}},
			}}}
			require.NoError(t, app.RunContext(t.Context(), append([]string{"easyp"}, tt.args...)))
			require.True(t, called)
		})
	}
}
