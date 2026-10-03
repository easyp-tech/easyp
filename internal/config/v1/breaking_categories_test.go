package v1

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
)

func TestBreakingCategoriesValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		categories string
		wantUse    []string
		wantError  bool
	}{
		{name: "omitted"},
		{name: "empty", categories: "  categories: []\n"},
		{name: "FILE", categories: "  categories: [FILE]\n", wantUse: []string{"FILE"}},
		{name: "duplicate FILE", categories: "  categories: [FILE, FILE]\n", wantUse: []string{"FILE", "FILE"}},
		{name: "PACKAGE", categories: "  categories: [PACKAGE]\n", wantUse: []string{"PACKAGE"}},
		{name: "WIRE_JSON", categories: "  categories: [WIRE_JSON]\n", wantUse: []string{"WIRE_JSON"}},
		{name: "WIRE", categories: "  categories: [WIRE]\n", wantUse: []string{"WIRE"}},
		{name: "mixed profiles", categories: "  categories: [WIRE, WIRE_JSON, PACKAGE, FILE]\n", wantUse: []string{"WIRE", "WIRE_JSON", "PACKAGE", "FILE"}},
		{name: "unknown", categories: "  categories: [UNKNOWN]\n", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := "version: v1\nbreaking:\n  baseline: git:main\n" + tt.categories
			assert.Equal(t, tt.wantError, config.HasErrors(ValidatePolicyYAML([]byte(raw))))
			path := filepath.Join(t.TempDir(), PolicyFile)
			require.NoError(t, os.WriteFile(path, []byte(raw), 0o644))
			issues, err := ValidateFile(path)
			require.NoError(t, err)
			assert.Equal(t, tt.wantError, config.HasErrors(issues))

			policy, err := ParsePolicy(strings.NewReader(raw))
			if tt.wantError {
				require.ErrorContains(t, err, "breaking.categories")
				return
			}
			require.NoError(t, err)
			cfg, err := policy.BreakingConfig("fallback")
			require.NoError(t, err)
			assert.Equal(t, tt.wantUse, cfg.Use)
			assert.Equal(t, "main", cfg.AgainstGitRef)
			if len(cfg.Use) > 0 {
				cfg.Use[0] = "changed"
				assert.Equal(t, tt.wantUse[0], policy.Breaking.Categories[0])
			}
		})
	}
}
