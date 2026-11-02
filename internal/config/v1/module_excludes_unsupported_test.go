package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseModuleRejectsExcludesDirective(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		text string
	}{
		{name: "directive", text: "excludes preview\n"},
		{name: "block", text: "excludes (\n  preview\n)\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseModule(strings.NewReader("module example.test/module\n" + tt.text))

			require.ErrorContains(t, err, "unknown")
			require.ErrorContains(t, err, "excludes")
		})
	}
}
