package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageSelectorValidation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		valid bool
	}{
		{name: "demo.v1", valid: true}, {name: "_demo.Type1", valid: true}, {name: "single", valid: true},
		{name: "demo.*"}, {name: "demo/one"}, {name: ".demo"}, {name: "demo..v1"}, {name: "demo-v1"}, {name: "demo.1"}, {name: ""}, {name: "demo."},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := "version: v1\ngenerate:\n  packages: [\"" + tt.name + "\"]\n"
			_, err := ParseGenerate(strings.NewReader(raw))
			assert.Equal(t, tt.valid, err == nil)
			if !tt.valid {
				require.ErrorContains(t, err, "generate.packages")
			}
		})
	}
	require.NoError(t, ValidatePackageSelectors(nil))
	require.NoError(t, ValidatePackageSelectors([]string{"demo.v1", "demo.v1"}))
}
