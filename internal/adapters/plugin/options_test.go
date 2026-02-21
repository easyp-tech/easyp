package plugin

import "testing"

func TestFlattenOptions(t *testing.T) {
	tests := []struct {
		name      string
		options   map[string][]string
		expected  string
		hasResult bool
	}{
		{
			name: "single scalar value",
			options: map[string][]string{
				"env": {"node"},
			},
			expected:  "env=node",
			hasResult: true,
		},
		{
			name: "multiple values for one key",
			options: map[string][]string{
				"outputServices": {"grpc-js", "generic-definitions"},
			},
			expected:  "outputServices=grpc-js,outputServices=generic-definitions",
			hasResult: true,
		},
		{
			name: "empty option value produces flag",
			options: map[string][]string{
				"useOptionals": {""},
			},
			expected:  "useOptionals",
			hasResult: true,
		},
		{
			name: "mixed options are deterministic by key and preserve value order",
			options: map[string][]string{
				"zeta": {"last"},
				"alpha": {
					"",
					"one",
				},
				"middle": {"x", ""},
			},
			expected:  "alpha,alpha=one,middle=x,middle,zeta=last",
			hasResult: true,
		},
		{
			name:      "empty map",
			options:   map[string][]string{},
			hasResult: false,
		},
		{
			name: "key with empty values list",
			options: map[string][]string{
				"empty": {},
			},
			hasResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := flattenOptions(tt.options)
			if ok != tt.hasResult {
				t.Fatalf("flattenOptions() has result = %v, want %v", ok, tt.hasResult)
			}
			if result != tt.expected {
				t.Fatalf("flattenOptions() result = %q, want %q", result, tt.expected)
			}
		})
	}
}
