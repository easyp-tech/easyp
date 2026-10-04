package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBreakingConfigDefaultsToFileProfile(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		`version: v1
breaking:
  baseline: git:main
`,
		`version: v1
breaking:
  baseline: git:main
  categories: []
`,
	} {
		policy, err := ParsePolicy(strings.NewReader(raw))
		require.NoError(t, err)

		cfg, err := policy.BreakingConfig("master")
		require.NoError(t, err)

		assert.Equal(t, []string{"FILE"}, cfg.Categories)
		assert.Equal(t, []string{"FILE"}, cfg.Use)
	}
}
