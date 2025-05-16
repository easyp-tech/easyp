package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DepAddressResolver_Resolve(t *testing.T) {
	mirrors := []MirrorConfig{
		{
			Origin: "github.com",
			Use:    "gitlab.com",
		},
	}

	resolver := NewDepAddressResolver(mirrors)

	tests := map[string]struct {
		requestedModule string
		expected        string
	}{
		"not replaced": {
			requestedModule: "vcs.com/github/octocat.git",
			expected:        "vcs.com/github/octocat.git",
		},
		"replaced": {
			requestedModule: "github.com/github/octocat.git",
			expected:        "gitlab.com/github/octocat.git",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res := resolver.Resolve(test.requestedModule)
			require.Equal(t, test.expected, res)
		})
	}
}
