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
	assertNoEmptyExpectedGotArtifact(t, issues)
}

func containsAny(s string, candidates ...string) bool {
	for _, candidate := range candidates {
		if candidate != "" && strings.Contains(s, candidate) {
			return true
		}
	}
	return false
}

func issuesBySeverityAndPath(issues []ValidationIssue, severity string, path string) []ValidationIssue {
	out := make([]ValidationIssue, 0)
	pathMarker := "(path: " + path + ")"
	for _, issue := range issues {
		if issue.Severity == severity && strings.Contains(issue.Message, pathMarker) {
			out = append(out, issue)
		}
	}
	return out
}

func assertNoEmptyExpectedGotArtifact(t *testing.T, issues []ValidationIssue) {
	t.Helper()
	for _, issue := range issues {
		require.NotContains(t, issue.Message, "expected , got", "unexpected malformed message: %q", issue.Message)
	}
}

func TestValidateRaw_PluginOptsRejectsNestedMapValue(t *testing.T) {
	content := `version: v1alpha
lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - directory: proto
  plugins:
    - command:
        - go
        - run
      out: .
      opts:
        l:
          L:
`

	issues, err := ValidateRaw([]byte(content))
	require.NoError(t, err)
	require.True(t, HasErrors(issues))
	assertNoEmptyExpectedGotArtifact(t, issues)

	var hasTypedError bool
	var hasNestedUnknown bool
	for _, issue := range issues {
		if issue.Severity == SeverityError &&
			strings.Contains(issue.Message, "opts value must be a scalar or sequence of scalars") &&
			strings.Contains(issue.Message, "expected scalar or sequence of scalars, got MappingNode") &&
			strings.Contains(issue.Message, "generate.plugins[0].opts.l") {
			hasTypedError = true
		}

		if issue.Severity == SeverityError &&
			strings.Contains(issue.Message, `unknown key "L"`) &&
			strings.Contains(issue.Message, "generate.plugins[0].opts.l.L") {
			hasNestedUnknown = true
		}
	}

	require.True(t, hasTypedError, "expected typed opts error with MappingNode, got issues: %#v", issues)
	require.True(t, hasNestedUnknown, "expected nested unknown key error for opts.l.L, got issues: %#v", issues)
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

func TestValidateRaw_DirectoryUnknownKey_NoDuplicate(t *testing.T) {
	content := `version: v1alpha
lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - directory:
        path: api
        pal: value
  plugins:
    - name: go
      out: .
`

	issues, err := ValidateRaw([]byte(content))
	require.NoError(t, err)
	require.False(t, HasErrors(issues), "directory unknown key should be warning-only, got: %#v", issues)
	assertNoEmptyExpectedGotArtifact(t, issues)

	warningsAtPath := issuesBySeverityAndPath(issues, SeverityWarn, "generate.inputs[0].directory.pal")

	require.Len(t, warningsAtPath, 1, "expected a single warning for directory unknown key, got: %#v", issues)
	require.Contains(t, warningsAtPath[0].Message, "unknown field under directory")
	require.NotContains(t, warningsAtPath[0].Message, `unknown key "pal"`)
}

func TestValidateRaw_GitRepoOut_IsUnknownKey(t *testing.T) {
	content := `version: v1alpha
lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - git_repo:
        url: github.com/acme/common@v1.0.0
        out: gen
  plugins:
    - name: go
      out: .
`

	issues, err := ValidateRaw([]byte(content))
	require.NoError(t, err)
	require.False(t, HasErrors(issues), "git_repo.out should be treated as unknown warning only, got: %#v", issues)
	assertNoEmptyExpectedGotArtifact(t, issues)

	warningsAtPath := issuesBySeverityAndPath(issues, SeverityWarn, "generate.inputs[0].git_repo.out")

	require.Len(t, warningsAtPath, 1, "expected one unknown warning for git_repo.out, got: %#v", issues)
	require.Contains(t, warningsAtPath[0].Message, `unknown key "out"`)
	require.Contains(t, warningsAtPath[0].Message, "(got ")
	require.NotContains(t, warningsAtPath[0].Message, "unknown field under directory")
}

func TestValidateRaw_MessageFormatting_GotOnly(t *testing.T) {
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
	assertNoEmptyExpectedGotArtifact(t, issues)

	var gotOnlyMessage string
	for _, issue := range issues {
		if issue.Severity == SeverityWarn && strings.Contains(issue.Message, `unknown key "unknown_field"`) {
			gotOnlyMessage = issue.Message
			break
		}
	}

	require.NotEmpty(t, gotOnlyMessage, "expected warning for unknown_field, got issues: %#v", issues)
	require.Contains(t, gotOnlyMessage, "(got ")
	require.NotContains(t, gotOnlyMessage, "(expected ")
}
