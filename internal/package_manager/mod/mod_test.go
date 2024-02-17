package mod

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterDirs(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected []string
	}{
		{
			name: "one dir",
			files: []string{
				"collect/file.proto",
				"collect/nested/file.proto",
			},
			expected: []string{
				"collect",
			},
		},
		{
			name: "several dir",
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

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := filterDirs(tc.files)
			require.ElementsMatch(t, tc.expected, result)
		})
	}
}
