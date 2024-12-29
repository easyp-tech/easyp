package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewModule(t *testing.T) {
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
				Version: Omitted,
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

func Test_RequestedVersion_GetParts(t *testing.T) {
	tests := map[string]struct {
		requestedVersion RequestedVersion
		expectedResult   GeneratedVersionParts
		expectedError    bool
	}{
		"not generated, simple tag": {
			requestedVersion: RequestedVersion("v1.2.3"),
			expectedResult:   GeneratedVersionParts{},
			expectedError:    true,
		},
		"not generated, tag with no `v` prefix": {
			requestedVersion: RequestedVersion("some_tag"),
			expectedResult:   GeneratedVersionParts{},
			expectedError:    true,
		},
		"not generated, with `-`": {
			requestedVersion: RequestedVersion("v1.2.3-rc"),
			expectedResult:   GeneratedVersionParts{},
			expectedError:    true,
		},
		"not generated, with several `-`": {
			requestedVersion: RequestedVersion("v1.2.3-rc-111222"),
			expectedResult:   GeneratedVersionParts{},
			expectedError:    true,
		},
		"Use Omitted": {
			requestedVersion: Omitted,
			expectedResult:   GeneratedVersionParts{},
			expectedError:    true,
		},
		"generated": {
			requestedVersion: "v0.0.0-20240222234643-814bf88cf225",
			expectedResult:   GeneratedVersionParts{Datetime: "20240222234643", CommitHash: "814bf88cf225"},
			expectedError:    false,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			result, err := tc.requestedVersion.GetParts()

			require.Equal(t, tc.expectedResult, result)

			if tc.expectedError {
				require.ErrorIs(t, err, ErrRequestedVersionNotGenerated)
				return
			}

			require.NoError(t, err)
		})
	}
}

func Test_RequestedVersion_IsGenerated(t *testing.T) {
	tests := map[string]struct {
		requestedVersion RequestedVersion
		expectedResult   bool
	}{
		"not generated, simple tag": {
			requestedVersion: RequestedVersion("v1.2.3"),
			expectedResult:   false,
		},
		"not generated, tag with no `v` prefix": {
			requestedVersion: RequestedVersion("some_tag"),
			expectedResult:   false,
		},
		"not generated, with `-`": {
			requestedVersion: RequestedVersion("v1.2.3-rc"),
			expectedResult:   false,
		},
		"not generated, with several `-`": {
			requestedVersion: RequestedVersion("v1.2.3-rc-111222"),
			expectedResult:   false,
		},
		"Use Omitted": {
			requestedVersion: Omitted,
			expectedResult:   false,
		},
		"generated": {
			requestedVersion: "v0.0.0-20240222234643-814bf88cf225",
			expectedResult:   true,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			result := tc.requestedVersion.IsGenerated()
			require.Equal(t, tc.expectedResult, result)
		})
	}
}

func Test_RequestedVersion_IsOmitted(t *testing.T) {
	tests := map[string]struct {
		requestedVersion RequestedVersion
		expectedResult   bool
	}{
		"not omitted": {
			requestedVersion: RequestedVersion("v1.2.3"),
			expectedResult:   false,
		},
		"omitted": {
			requestedVersion: Omitted,
			expectedResult:   true,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			result := tc.requestedVersion.IsOmitted()
			require.Equal(t, tc.expectedResult, result)
		})
	}
}

func Test_GeneratedVersionParts_GetVersionString(t *testing.T) {
	tests := map[string]struct {
		parts          GeneratedVersionParts
		expectedResult string
	}{
		"case 1": {
			parts:          GeneratedVersionParts{Datetime: "20240222234643", CommitHash: "814bf88cf225"},
			expectedResult: "v0.0.0-20240222234643-814bf88cf225",
		},
		"case 2": {
			parts:          GeneratedVersionParts{Datetime: "20230212224650", CommitHash: "914af88cf235"},
			expectedResult: "v0.0.0-20230212224650-914af88cf235",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			result := tc.parts.GetVersionString()
			require.Equal(t, tc.expectedResult, result)
		})
	}
}
