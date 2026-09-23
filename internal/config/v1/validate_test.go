package v1

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateVersionlessV1Policy(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "easyp.yaml")
	err := os.WriteFile(path, []byte("linters:\n  default: MINIMAL\nbreaking:\n  baseline: git:main\n"), 0o644)
	require.NoError(t, err)

	issues, err := ValidateFile(path)
	require.NoError(t, err)
	assert.Empty(t, issues)
}

func TestValidateRejectsLegacyPolicy(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "easyp.yaml")
	require.NoError(t, os.WriteFile(path, []byte("lint:\n  use: [MINIMAL]\n"), 0o644))
	issues, err := ValidateFile(path)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "yaml_validation", issues[0].Code)
	require.Positive(t, issues[0].Line)
	require.Positive(t, issues[0].Column)
}

func TestValidateV1Lock(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		content string
		valid   bool
	}{
		{name: "empty graph", content: "version: 1\nmodules: []\n", valid: true},
		{name: "wrong version", content: "version: 2\nmodules: []\n"},
		{name: "incomplete module", content: "version: 1\nmodules:\n  - source: example.com/dep\n    version: v1.0.0\n"},
		{name: "bad hash", content: "version: 1\nmodules:\n  - source: example.com/dep\n    version: v1.0.0\n    commit: 0000000000000000000000000000000000000000\n    hash: h1:bad\n"},
		{name: "unknown field", content: "version: 1\nmodules: []\nunknown: true\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "protobuf.lock")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), 0o644))
			issues, err := ValidateFile(path)
			require.NoError(t, err)
			if tc.valid {
				assert.Empty(t, issues)
			} else {
				require.Len(t, issues, 1)
				assert.Equal(t, "v1_validation", issues[0].Code)
			}
		})
	}
}

func TestValidatePolicyReportsMultipleLocatedYAMLIssues(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "easyp.yaml")
	content := "linters:\n  enable: INVALID\nissues:\n  exclude-rules:\n    - linters: [GOOD]\n      extra: true\nbreaking:\n  ignore: wrong\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	issues, err := ValidateFile(path)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(issues), 3)
	for _, issue := range issues {
		assert.Equal(t, "yaml_validation", issue.Code)
		assert.Equal(t, "error", issue.Severity)
		assert.Positive(t, issue.Line)
		assert.Positive(t, issue.Column)
	}
}

func TestValidateGenerateReportsMultipleLocatedYAMLIssues(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "easyp.gen.yaml")
	content := "plugins:\n  - name: python\n    out: gen\n    opts: 17\n  - name: go\n    out: gen/go\n    unknown: true\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	issues, err := ValidateFile(path)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(issues), 2)
	for _, issue := range issues {
		assert.Equal(t, "yaml_validation", issue.Code)
		assert.Equal(t, "error", issue.Severity)
		assert.Positive(t, issue.Line)
		assert.Positive(t, issue.Column)
	}
}

func TestValidateGenerateAcceptsManagedAndPluginOptions(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "easyp.gen.yaml")
	content := "version: v1\ngenerate:\n  managed:\n    enabled: true\n    disable:\n      - field_option: json_name\n        field: example.v1.Message.name\n    override:\n      - file_option: go_package_prefix\n        value: example.com/gen\nplugins:\n  - name: python\n    out: gen/python\n    opts:\n      foo: bar\n      paths: [source_relative]\n  - name: go\n    out: gen/go\n    opts: [paths=source_relative]\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	issues, err := ValidateFile(path)
	require.NoError(t, err)
	assert.Empty(t, issues)
}

func TestValidatePolicyAcceptsV1Fields(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "easyp.yaml")
	content := "version: v1\nlinters:\n  default: STANDARD\n  enable: [ENUM_ZERO_VALUE_SUFFIX]\nlinters-settings:\n  ENUM_ZERO_VALUE_SUFFIX:\n    suffix: UNSPECIFIED\nissues:\n  exclude-rules:\n    - linters: [SERVICE_SUFFIX]\nbreaking:\n  baseline: git:main\n  ignore: [generated]\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	issues, err := ValidateFile(path)
	require.NoError(t, err)
	assert.Empty(t, issues)
}
