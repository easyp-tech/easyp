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
