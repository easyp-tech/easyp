package mod

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterOnlyProtoDirs(t *testing.T) {
	tests := map[string]struct {
		files    []string
		expected []string
	}{
		"one dir": {
			files: []string{
				"collect/file.proto",
				"collect/nested/file.proto",
			},
			expected: []string{
				"collect",
			},
		},
		"several dir": {
			files: []string{
				"collect/file.proto",
				"collect/nested/file.proto",
				"storage/file.proto",
				"storage/111/file.proto",
				"storage/222/file.proto",
				"wo_proto/file.txt",
			},
			expected: []string{
				"collect",
				"storage",
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			result := filterOnlyProtoDirs(tc.files)
			require.ElementsMatch(t, tc.expected, result)
		})
	}
}
