package core

import (
	"testing"
)

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		prefix   string
		expected string
	}{
		{
			name:     "simple relative path",
			path:     "proto/task/v1/task.proto",
			prefix:   "proto",
			expected: "task/v1/task.proto",
		},
		{
			name:     "dotslash prefix",
			path:     "proto/common/v1/common.proto",
			prefix:   "./proto",
			expected: "common/v1/common.proto",
		},
		{
			name:     "nested directories",
			path:     "api/proto/v1/service.proto",
			prefix:   "api/proto",
			expected: "v1/service.proto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripPrefix(tt.path, tt.prefix)
			if result != tt.expected {
				t.Errorf("stripPrefix(%q, %q) = %q, want %q",
					tt.path, tt.prefix, result, tt.expected)
			}
		})
	}
}
