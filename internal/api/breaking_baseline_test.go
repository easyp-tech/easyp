package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestBreakingBaselinePrecedence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, baseline string
		args           []string
		want           string
		wantErr        bool
	}{
		{name: "policy before default", baseline: "git:release", want: "release"},
		{name: "explicit before policy", baseline: "git:release", args: []string{"--against", "main"}, want: "main"},
		{name: "explicit value equals default", baseline: "git:release", args: []string{"--against", "master"}, want: "master"},
		{name: "default without policy", want: "master"},
		{name: "explicit without policy", args: []string{"--against", "main"}, want: "main"},
		{name: "empty explicit rejected", baseline: "git:release", args: []string{"--against", ""}, wantErr: true},
		{name: "whitespace explicit rejected", baseline: "git:release", args: []string{"--against", "   "}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			policy := v1.Policy{Breaking: v1.BreakingPolicy{Baseline: tt.baseline}}
			// urfave uses a process-global help flag; these independent parsers do not need it.
			app := &cli.App{
				HideHelp:    true,
				HideVersion: true,
				Flags:       []cli.Flag{&cli.StringFlag{Name: flagAgainstBranchName.Name, Value: flagAgainstBranchName.Value}},
				Action: func(ctx *cli.Context) error {
					cfg, err := resolveV1BreakingConfig(ctx, policy)
					if err != nil {
						return err
					}
					assert.Equal(t, tt.want, cfg.AgainstGitRef)
					return nil
				},
			}
			err := app.Run(append([]string{"test"}, tt.args...))
			if tt.wantErr {
				require.ErrorContains(t, err, "--against")
			} else {
				require.NoError(t, err)
			}
		})
	}
	assert.False(t, flagAgainstBranchName.Required)
	assert.False(t, flagAgainstBranchName.HasBeenSet)
}
