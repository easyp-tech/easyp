package storage

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core/models"
)

func TestGetRenamer(t *testing.T) {
	tests := map[string]struct {
		moduleConfig   models.ModuleConfig
		passedFile     string
		expectedResult string
	}{
		"directories are empty": {
			moduleConfig: models.ModuleConfig{
				Directories: nil,
			},
			passedFile:     "proto/file.proto",
			expectedResult: "proto/file.proto",
		},
		"directories contain one dir": {
			moduleConfig: models.ModuleConfig{
				Directories: []string{"proto/protovalidate"},
			},
			passedFile:     "proto/protovalidate/buf/validate/validate.proto",
			expectedResult: "buf/validate/validate.proto",
		},
		"directories contain several dirs": {
			moduleConfig: models.ModuleConfig{
				Directories: []string{"proto/protovalidate", "proto/protovalidate-testing"},
			},
			passedFile:     "proto/protovalidate/buf/validate/validate.proto",
			expectedResult: "buf/validate/validate.proto",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			renamer := getRenamer(tc.moduleConfig)
			result := renamer(tc.passedFile)
			require.Equal(t, tc.expectedResult, result)
		})
	}
}
