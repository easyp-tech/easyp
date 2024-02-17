package dependency

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDependency(t *testing.T) {
	tests := []struct {
		name           string
		module         string
		expectedResult Dependency
	}{
		{
			name:   "with version",
			module: "github.com/company/repository@v1.2.3",
			expectedResult: Dependency{
				Name:    "github.com/company/repository",
				Version: "v1.2.3",
			},
		},
		{
			name:   "without version",
			module: "github.com/company/repository",
			expectedResult: Dependency{
				Name:    "github.com/company/repository",
				Version: "",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := ParseDependency(tc.module)
			require.Equal(t, tc.expectedResult, result)
		})
	}
}
