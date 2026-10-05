package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseModuleRejectsNormalizedRootEscapes(t *testing.T) {
	t.Parallel()

	for _, root := range []string{"..", "foo/../..", "foo/../../bar"} {
		root := root
		t.Run(root, func(t *testing.T) {
			t.Parallel()

			_, err := ParseModule(strings.NewReader("module example.test/root\nroots " + root + "\n"))

			require.ErrorContains(t, err, "leaves the module")
			assert.Contains(t, err.Error(), root)
		})
	}
}

func TestParseModuleStoresCleanLocalRoots(t *testing.T) {
	t.Parallel()

	module, err := ParseModule(strings.NewReader("module example.test/root\nroots foo/../bar\n"))

	require.NoError(t, err)
	assert.Equal(t, []string{"bar"}, module.Roots)
}
