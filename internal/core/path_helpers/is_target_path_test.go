package path_helpers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core/path_helpers"
)

func Test_IsTargetPath(t *testing.T) {
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
		"target only dirs (target dir match) more complex 2": {
			targetPath: "collections/v1",
			filePath:   "collections/v1",
			expected:   true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res := path_helpers.IsTargetPath(test.targetPath, test.filePath)
			require.Equal(t, test.expected, res)
		})
	}
}

func Test_IsIgnoredPath(t *testing.T) {
	tests := map[string]struct {
		targetPath string
		ignore     []string
		expected   bool
	}{
		"ignore is empty (dir)": {
			targetPath: "some/path",
			ignore:     nil,
			expected:   false,
		},
		"ignore is empty (file)": {
			targetPath: "some/path/file.json",
			ignore:     nil,
			expected:   false,
		},
		"ignore is filled (file)": {
			targetPath: "some/path/file.json",
			ignore:     []string{"path", "another"},
			expected:   false,
		},
		"ignored (dir)": {
			targetPath: "some/path",
			ignore:     []string{"some"},
			expected:   true,
		},
		"2*ignored (dir)": {
			targetPath: "some/path",
			ignore:     []string{"some", "path"},
			expected:   true,
		},
		"2.1*ignored (dir)": {
			targetPath: "path",
			ignore:     []string{"some", "path"},
			expected:   true,
		},
		"ignored (dir/dir)": {
			targetPath: "some/path",
			ignore:     []string{"some/path"},
			expected:   true,
		},
		"ignored (file)": {
			targetPath: "some/path/file.json",
			ignore:     []string{"some"},
			expected:   true,
		},
		"ignored (dir/dir/file)": {
			targetPath: "some/path/file.json",
			ignore:     []string{"some/path"},
			expected:   true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res := path_helpers.IsIgnoredPath(test.targetPath, test.ignore)
			require.Equal(t, test.expected, res)
		})
	}
}
