package v1

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
)

func TestValidateFileAcceptsValidConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		contents string
	}{
		{
			name:     "versionless policy",
			filename: PolicyFile,
			contents: "linters:\n  default: MINIMAL\nbreaking:\n  baseline: git:main\n",
		},
		{
			name:     "custom policy filename",
			filename: "project.yaml",
			contents: "version: v1\nlinters:\n  default: MINIMAL\n",
		},
		{
			name:     "policy settings and pathless exclusions",
			filename: PolicyFile,
			contents: `version: v1
linters:
  default: STANDARD
  enable: [ENUM_ZERO_VALUE_SUFFIX]
linters-settings:
  ENUM_ZERO_VALUE_SUFFIX:
    suffix: UNSPECIFIED
issues:
  exclude-rules:
    - linters: [SERVICE_SUFFIX]
breaking:
  baseline: git:main
  ignore: [generated]
`,
		},
		{
			name:     "managed rules and both plugin option formats",
			filename: GenerateFile,
			contents: `version: v1
generate:
  managed:
    enabled: true
    disable:
      - field_option: json_name
        field: example.v1.Message.name
    override:
      - file_option: go_package_prefix
        value: example.com/gen
plugins:
  - name: python
    out: gen/python
    opts:
      foo: bar
      paths: [source_relative]
  - name: go
    out: gen/go
    opts: [paths=source_relative]
`,
		},
		{
			name:     "module manifest",
			filename: ModuleFile,
			contents: "module example.com/service\nroots (\n  proto\n)\n",
		},
		{
			name:     "plugin binary path",
			filename: GenerateFile,
			contents: "version: v1\nplugins:\n  - path: ./tools/protoc-gen-custom\n    out: gen\n",
		},
		{
			name:     "custom plugin command",
			filename: GenerateFile,
			contents: "version: v1\nplugins:\n  - command: [sh, ./tools/run-plugin]\n    out: gen\n",
		},
		{
			name:     "empty dependency graph",
			filename: LockFile,
			contents: "version: 1\nmodules: []\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), tt.filename)
			require.NoError(t, os.WriteFile(path, []byte(tt.contents), 0o644))

			issues, err := ValidateFile(path)

			require.NoError(t, err)
			assert.Empty(t, issues)
		})
	}
}

func TestValidateFileReportsLocatedYAMLIssues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		filename      string
		contents      string
		expectedCount int
	}{
		{
			name:          "legacy policy",
			filename:      PolicyFile,
			contents:      "lint:\n  use: [MINIMAL]\n",
			expectedCount: 1,
		},
		{
			name:     "multiple policy issues",
			filename: PolicyFile,
			contents: `linters:
  enable: INVALID
issues:
  exclude-rules:
    - linters: [GOOD]
      extra: true
breaking:
  ignore: wrong
`,
			expectedCount: 3,
		},
		{
			name:     "multiple generator issues",
			filename: GenerateFile,
			contents: `plugins:
  - name: python
    out: gen
    opts: 17
  - name: go
    out: gen/go
    unknown: true
`,
			expectedCount: 2,
		},
		{
			name:          "nested plugin option issue",
			filename:      GenerateFile,
			contents:      "plugins:\n  - name: python\n    out: gen\n    opts:\n      invalid:\n        nested: value\n",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), tt.filename)
			require.NoError(t, os.WriteFile(path, []byte(tt.contents), 0o644))

			issues, err := ValidateFile(path)

			require.NoError(t, err)
			require.Len(t, issues, tt.expectedCount)
			for _, issue := range issues {
				assert.Equal(t, "yaml_validation", issue.Code)
				assert.Equal(t, config.SeverityError, issue.Severity)
				assert.Positive(t, issue.Line)
				assert.Positive(t, issue.Column)
			}
		})
	}
}

func TestValidateFileUsesGeneratedSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		contents string
		path     string
	}{
		{
			name:     "unknown linter setting",
			filename: PolicyFile,
			contents: "linters-settings:\n  UNKNOWN:\n    suffix: BAD\n",
			path:     `["linters-settings"].UNKNOWN`,
		},
		{
			name:     "invalid baseline",
			filename: PolicyFile,
			contents: "breaking:\n  baseline: main\n",
			path:     "breaking.baseline",
		},
		{
			name:     "plugin without identity",
			filename: GenerateFile,
			contents: "plugins:\n  - out: gen\n",
			path:     "plugins[0]",
		},
		{
			name:     "managed override without value",
			filename: GenerateFile,
			contents: "generate:\n  managed:\n    override:\n      - file_option: go_package_prefix\n",
			path:     "generate.managed.override[0].value",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), tt.filename)
			require.NoError(t, os.WriteFile(path, []byte(tt.contents), 0o644))

			issues, err := ValidateFile(path)

			require.NoError(t, err)
			require.NotEmpty(t, issues)
			assert.Equal(t, "yaml_validation", issues[0].Code)
			assert.Contains(t, issues[0].Message, tt.path)
			assert.Positive(t, issues[0].Line)
			assert.Positive(t, issues[0].Column)
		})
	}
}

func TestValidateFileReportsSemanticIssues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		filename        string
		contents        string
		expectedMessage string
	}{
		{
			name:            "unsupported lock version",
			filename:        LockFile,
			contents:        "version: 2\nmodules: []\n",
			expectedMessage: "unsupported protobuf.lock version 2",
		},
		{
			name:            "incomplete locked module",
			filename:        LockFile,
			contents:        "version: 1\nmodules:\n  - source: example.com/dep\n    version: v1.0.0\n",
			expectedMessage: "incomplete lock entry",
		},
		{
			name:            "invalid content hash",
			filename:        LockFile,
			contents:        "version: 1\nmodules:\n  - source: example.com/dep\n    version: v1.0.0\n    commit: 0000000000000000000000000000000000000000\n    hash: h1:bad\n",
			expectedMessage: "invalid content hash",
		},
		{
			name:            "unknown lock field",
			filename:        LockFile,
			contents:        "version: 1\nmodules: []\nunknown: true\n",
			expectedMessage: "field unknown not found",
		},
		{
			name:            "missing module identity",
			filename:        ModuleFile,
			contents:        "roots (\n  proto\n)\n",
			expectedMessage: "missing module directive",
		},
		{
			name:            "unsupported linter inheritance",
			filename:        PolicyFile,
			contents:        "linters:\n  extends: shared.yaml\n",
			expectedMessage: "linters.extends policy loading is not implemented",
		},
		{
			name:            "unsupported breaking categories",
			filename:        PolicyFile,
			contents:        "breaking:\n  categories: [WIRE]\n",
			expectedMessage: "breaking.categories",
		},
		{
			name:            "unpinned remote plugin",
			filename:        GenerateFile,
			contents:        "plugins:\n  - remote: example.com/go\n    out: gen\n",
			expectedMessage: "remote plugin requires a pinned semantic version",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), tt.filename)
			require.NoError(t, os.WriteFile(path, []byte(tt.contents), 0o644))

			issues, err := ValidateFile(path)

			require.NoError(t, err)
			require.Len(t, issues, 1)
			assert.Equal(t, "v1_validation", issues[0].Code)
			assert.Equal(t, config.SeverityError, issues[0].Severity)
			assert.Contains(t, issues[0].Message, tt.expectedMessage)
		})
	}
}

func TestValidateFileReadError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
	}{
		{name: "missing policy", filename: PolicyFile},
		{name: "missing generator", filename: GenerateFile},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), tt.filename)

			issues, err := ValidateFile(path)

			require.ErrorIs(t, err, os.ErrNotExist)
			assert.Nil(t, issues)
			var pathErr *os.PathError
			require.ErrorAs(t, err, &pathErr)
			assert.Equal(t, path, pathErr.Path)
		})
	}
}
