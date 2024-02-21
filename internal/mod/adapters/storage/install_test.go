package storage

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

func TestGetRenamer(t *testing.T) {
	tests := []struct {
		name           string
		moduleConfig   models.ModuleConfig
		passedFile     string
		expectedResult string
	}{
		{
			name: "Directories are empty",
			moduleConfig: models.ModuleConfig{
				Directories: nil,
			},
			passedFile:     "proto/file.proto",
			expectedResult: "proto/file.proto",
		},
		{
			name: "Directories contain one dir",
			moduleConfig: models.ModuleConfig{
				Directories: []string{"proto/protovalidate"},
			},
			passedFile:     "proto/protovalidate/buf/validate/validate.proto",
			expectedResult: "buf/validate/validate.proto",
		},
		{
			name: "Directories contain several dirs",
			moduleConfig: models.ModuleConfig{
				Directories: []string{"proto/protovalidate", "proto/protovalidate-testing"},
			},
			passedFile:     "proto/protovalidate/buf/validate/validate.proto",
			expectedResult: "buf/validate/validate.proto",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			renamer := getRenamer(tc.moduleConfig)
			result := renamer(tc.passedFile)
			require.Equal(t, tc.expectedResult, result)
		})
	}
}
