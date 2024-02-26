package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	tests := map[string]struct {
		dependency     string
		expectedResult Module
	}{
		"with version": {
			dependency: "github.com/company/repository@v1.2.3",
			expectedResult: Module{
				Name:    "github.com/company/repository",
				Version: "v1.2.3",
			},
		},
		"without version": {
			dependency: "github.com/company/repository",
			expectedResult: Module{
				Name:    "github.com/company/repository",
				Version: "",
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			result := NewModule(tc.dependency)
			require.Equal(t, tc.expectedResult, result)
		})
	}
}
