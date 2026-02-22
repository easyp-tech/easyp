package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRaw_PluginOptsSupportsScalarAndSequence(t *testing.T) {
	content := `version: v1alpha
lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - directory: proto
  plugins:
    - name: go
      out: .
      opts:
        env: node
        outputServices:
          - grpc-js
          - generic-definitions
`

	issues, err := ValidateRaw([]byte(content))
	require.NoError(t, err)
	require.False(t, HasErrors(issues))
}

func TestValidateRaw_PluginOptsRejectsNestedValues(t *testing.T) {
	content := `version: v1alpha
lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - directory: proto
  plugins:
    - name: go
      out: .
      opts:
        outputServices:
          - grpc-js
          - nested:
              key: value
`

	issues, err := ValidateRaw([]byte(content))
	require.NoError(t, err)
	require.True(t, HasErrors(issues))

	found := false
	for _, issue := range issues {
		if issue.Severity != SeverityError {
			continue
		}
		if issue.Message == "" {
			continue
		}
		if containsAny(issue.Message, "opts array item must be a scalar value", "opts value must be a scalar or sequence of scalars") {
			found = true
			break
		}
	}
	require.True(t, found, "expected plugin opts validation error, got issues: %#v", issues)
}

func containsAny(s string, candidates ...string) bool {
	for _, candidate := range candidates {
		if candidate != "" && strings.Contains(s, candidate) {
			return true
		}
	}
	return false
}

func TestValidateRaw_BreakingSchemaValidKeys(t *testing.T) {
	content := `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
breaking:
  ignore:
    - proto/legacy
  against_git_ref: main
`

	issues, err := ValidateRaw([]byte(content))
	require.NoError(t, err)
	require.False(t, HasErrors(issues), "valid breaking section should not produce errors, got: %v", issues)
}

func TestValidateRaw_BreakingSchemaUnknownKey(t *testing.T) {
	content := `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
breaking:
  ignore:
    - proto/legacy
  unknown_field: true
`

	issues, err := ValidateRaw([]byte(content))
	require.NoError(t, err)

	// Should produce at least one warning for the unknown key.
	var hasWarning bool
	for _, issue := range issues {
		if issue.Severity == SeverityWarn && strings.Contains(issue.Message, "unknown_field") {
			hasWarning = true
			break
		}
	}
	require.True(t, hasWarning, "expected warning for unknown key 'unknown_field' in breaking section, got: %v", issues)
}
