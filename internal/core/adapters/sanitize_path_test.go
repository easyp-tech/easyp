package adapters_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/mod/adapters"
)

func Test_SanitizePath(t *testing.T) {
	tests := map[string]struct {
		source string
		expect string
	}{
		"with slashes": {
			source: "dir/version",
			expect: "dir-version",
		},
		"without slashes": {
			source: "version",
			expect: "version",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := adapters.SanitizePath(tc.source)

			require.Equal(t, tc.expect, res)
		})
	}
}
