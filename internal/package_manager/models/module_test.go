package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	tests := []struct {
		name           string
		dependency     string
		expectedResult Module
	}{
		{
			name:       "with version",
			dependency: "github.com/company/repository@v1.2.3",
			expectedResult: Module{
				Name:    "github.com/company/repository",
				Version: "v1.2.3",
			},
		},
		{
			name:       "without version",
			dependency: "github.com/company/repository",
			expectedResult: Module{
				Name:    "github.com/company/repository",
				Version: "",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := NewModule(tc.dependency)
			require.Equal(t, tc.expectedResult, result)
		})
	}
}
