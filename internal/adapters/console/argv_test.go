package console

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandArgvHelper(t *testing.T) {
	if os.Getenv("EASYP_ARGV_TEST_HELPER") != "1" {
		return
	}
	var args []string
	for i, arg := range os.Args {
		if arg == "--" {
			args = os.Args[i+1:]
			break
		}
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(2)
	}
	body, err := json.Marshal(struct {
		Args  []string
		Stdin string
	}{args, string(data)})
	if err != nil {
		os.Exit(2)
	}
	fmt.Print(string(body))
	os.Exit(0)
}

func TestRunCmdWithStdinPreservesArgv(t *testing.T) {
	// The helper uses an environment flag; do not share it with parallel tests.
	t.Setenv("EASYP_ARGV_TEST_HELPER", "1")
	executable, err := os.Executable()
	require.NoError(t, err)
	args := []string{"argument with spaces", "'quoted'", "", "$HOME", "$(printf should-not-run)", "; echo not-a-command", "*"}
	for _, tt := range []struct {
		name string
		run  func(context.Context, string, io.Reader, string, ...string) (string, error)
	}{
		{name: "bash adapter", run: (bash{}).RunCmdWithStdin},
		{name: "powershell adapter", run: (powershell{}).RunCmdWithStdin},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := tt.run(t.Context(), t.TempDir(), strings.NewReader("protobuf input\x00"), executable,
				append([]string{"-test.run=^TestCommandArgvHelper$", "--"}, args...)...)
			require.NoError(t, err)
			var got struct {
				Args  []string
				Stdin string
			}
			require.NoError(t, json.Unmarshal([]byte(out), &got))
			require.Equal(t, args, got.Args)
			require.Equal(t, "protobuf input\x00", got.Stdin)
		})
	}
}
