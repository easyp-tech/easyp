package flags

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestGetFormatRespectsExplicitScope(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name                string
		global, local, want string
	}{
		{name: "command default", want: "json"},
		{name: "explicit global text", global: "text", want: "text"},
		{name: "local overrides global", global: "text", local: "json", want: "json"},
		{name: "explicit local text", local: "text", want: "text"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			global := flag.NewFlagSet("global", flag.ContinueOnError)
			global.String("format", "text", "")
			local := flag.NewFlagSet("command", flag.ContinueOnError)
			local.String("format", "text", "")
			if tt.global != "" {
				require.NoError(t, global.Set("format", tt.global))
			}
			if tt.local != "" {
				require.NoError(t, local.Set("format", tt.local))
			}
			parent := cli.NewContext(&cli.App{}, global, nil)
			child := cli.NewContext(parent.App, local, parent)
			require.Equal(t, tt.want, GetFormat(child, "json"))
		})
	}
}
