package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestRewriteV1RequiredVersionsBlockWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opening string
	}{
		{name: "single space", opening: "require ("},
		{name: "no space", opening: "require("},
		{name: "tab", opening: "require\t("},
		{name: "multiple spaces", opening: "require   ("},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := "module example.com/root\n" + tt.opening + " // dependencies\n\thttps://example.com/common.git v1.0.0 // indirect\n)\n"
			_, err := v1.ParseModule(strings.NewReader(original))
			require.NoError(t, err)

			updated, err := rewriteV1RequiredVersions([]byte(original), map[string]string{
				"https://example.com/common.git": "v1.1.0",
			})
			require.NoError(t, err)
			assert.Equal(t, strings.Replace(original, "v1.0.0", "v1.1.0", 1), string(updated))
		})
	}
}
