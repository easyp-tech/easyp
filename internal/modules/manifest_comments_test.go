package modules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestManifestEditsPreserveHashComments(t *testing.T) {
	t.Parallel()
	original := "\ufeffmodule example.com/app\nrequire (\n  # ignored-module v9.0.0\n  https://example.com/dep v1.0.0 # keep this comment\n)\n"
	parsed, err := v1.ParseModule(strings.NewReader(original))
	require.NoError(t, err)
	require.Equal(t, parsed.Requires, directV1Module([]byte(original), parsed).Requires)
	got := rewriteV1RequiredVersions([]byte(original), map[string]string{"https://example.com/dep": "v1.1.0"})
	require.Equal(t, strings.Replace(original, "v1.0.0", "v1.1.0", 1), string(got))
}
