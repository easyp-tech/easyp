package storage

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core/models"
)

func Test_isVersionsMatched(t *testing.T) {
	tests := map[string]struct {
		requestedVersion models.RequestedVersion
		lockFileVersion  string
		expectedResult   bool
	}{
		"requested version is omitted": {
			requestedVersion: models.Omitted,
			lockFileVersion:  gofakeit.Word(),
			expectedResult:   true,
		},
		"requested version and lock file are matched": {
			requestedVersion: "v1.2.3",
			lockFileVersion:  "v1.2.3",
			expectedResult:   true,
		},
		"requested version and lock file are not matched": {
			requestedVersion: "v1.2.3-1",
			lockFileVersion:  "v1.2.3",
			expectedResult:   false,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			result := isVersionsMatched(tc.requestedVersion, tc.lockFileVersion)
			require.Equal(t, tc.expectedResult, result)
		})
	}
}
