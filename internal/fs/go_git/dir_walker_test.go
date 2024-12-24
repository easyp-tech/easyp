package go_git

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_isTargetFile(t *testing.T) {
	tests := map[string]struct {
		targetPath string
		filePath   string
		expected   bool
	}{
		"not target (in root)": {
			targetPath: "proto_files",
			filePath:   "message.proto",
			expected:   false,
		},
		"not target (different dir)": {
			targetPath: "proto_files",
			filePath:   "collections/message.proto",
			expected:   false,
		},
		"target (root dir is target dir)": {
			targetPath: ".",
			filePath:   "collections/message.proto",
			expected:   true,
		},
		"target (target dir match)": {
			targetPath: "collections",
			filePath:   "collections/message.proto",
			expected:   true,
		},
		"target (target dir match) more complex": {
			targetPath: "collections",
			filePath:   "collections/v1/message.proto",
			expected:   true,
		},
		"target (target dir match) more complex 2": {
			targetPath: "collections/v1",
			filePath:   "collections/v1/message.proto",
			expected:   true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res := isTargetFile(test.targetPath, test.filePath)
			require.Equal(t, test.expected, res)
		})
	}
}
