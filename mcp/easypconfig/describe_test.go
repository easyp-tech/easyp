package easypconfig

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestDescribeV1Files(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      DescribeInput
		wantFile   string
		wantPath   string
		wantField  string
		wantSchema bool
	}{
		{name: "default policy", input: DescribeInput{Path: "$.linters.default"}, wantFile: v1.PolicyFile, wantPath: "linters.default", wantField: "linters.default", wantSchema: true},
		{name: "generator plugin", input: DescribeInput{File: v1.GenerateFile, Path: "plugins[3].out"}, wantFile: v1.GenerateFile, wantPath: "plugins[].out", wantField: "plugins[].out", wantSchema: true},
		{name: "plugin binary path", input: DescribeInput{File: v1.GenerateFile, Path: "plugins[0].path"}, wantFile: v1.GenerateFile, wantPath: "plugins[].path", wantField: "plugins[].path", wantSchema: true},
		{name: "custom plugin command", input: DescribeInput{File: v1.GenerateFile, Path: "plugins[0].command"}, wantFile: v1.GenerateFile, wantPath: "plugins[].command", wantField: "plugins[].command", wantSchema: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Describe(tt.input)

			require.NoError(t, err)
			assert.Equal(t, "v1", got.SchemaVersion)
			assert.Equal(t, tt.wantFile, got.File)
			assert.Equal(t, tt.wantPath, got.SelectedPath)
			assert.Contains(t, fieldPaths(got.Fields), tt.wantField)
			if tt.wantSchema {
				assert.NotEmpty(t, got.Schema)
			} else {
				assert.Empty(t, got.Schema)
			}
		})
	}
}

func TestDescribeRejectsUnknownSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input DescribeInput
		want  string
	}{
		{name: "unknown file", input: DescribeInput{File: "buf.yaml"}, want: "unknown config file"},
		{name: "module manifest is outside MCP", input: DescribeInput{File: v1.ModuleFile}, want: "unknown config file"},
		{name: "lock file is outside MCP", input: DescribeInput{File: v1.LockFile}, want: "unknown config file"},
		{name: "removed v0 path", input: DescribeInput{File: v1.GenerateFile, Path: "generate.inputs"}, want: "unknown path"},
		{name: "unknown policy field", input: DescribeInput{Path: "lint.use"}, want: "unknown path"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := Describe(tt.input)

			require.ErrorContains(t, err, tt.want)
		})
	}
}

func TestDescribeNotesForReservedFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file string
		path string
		want string
	}{
		{name: "package filter", file: v1.GenerateFile, path: "generate.packages", want: "generation fails"},
		{name: "linter inheritance", file: v1.PolicyFile, path: "linters.extends", want: "not implemented"},
		{name: "issue path matching", file: v1.PolicyFile, path: "issues.exclude-rules[0].path", want: "not implemented"},
		{name: "breaking categories", file: v1.PolicyFile, path: "breaking.categories", want: "not supported"},
		{name: "breaking unstable", file: v1.PolicyFile, path: "breaking.ignore_unstable", want: "not supported"},
		{name: "breaking inheritance", file: v1.PolicyFile, path: "breaking.extends", want: "not implemented"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Describe(DescribeInput{File: tt.file, Path: tt.path})

			require.NoError(t, err)
			require.Len(t, got.Fields, 1)
			assert.Contains(t, strings.Join(got.Notes, " "), tt.want)
			assert.Contains(t, strings.Join(got.Fields[0].Notes, " "), tt.want)
		})
	}
}

func TestDescribeUsesGeneratedSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file string
		key  string
	}{
		{name: "policy", file: v1.PolicyFile, key: "easyp"},
		{name: "generator", file: v1.GenerateFile, key: "easyp.gen"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Describe(DescribeInput{File: tt.file})
			require.NoError(t, err)
			raw, err := v1.SchemaJSON(tt.key)
			require.NoError(t, err)
			var want map[string]any
			require.NoError(t, json.Unmarshal(raw, &want))

			assert.Equal(t, want, got.Schema)
		})
	}
}

func TestFieldDescriptionsMatchSchema(t *testing.T) {
	t.Parallel()

	for _, file := range []string{v1.PolicyFile, v1.GenerateFile} {
		file := file
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			index, err := SchemaByPathFor(file)

			require.NoError(t, err)
			for path := range descriptions[file] {
				assert.Contains(t, index, path)
			}
		})
	}
}

func TestDescribeOptionsAndIndependentResults(t *testing.T) {
	t.Parallel()

	falseValue := false
	withoutDetails, err := Describe(DescribeInput{File: v1.GenerateFile, Path: "plugins", IncludeSchema: &falseValue, IncludeFields: &falseValue, IncludeExamples: &falseValue})
	require.NoError(t, err)
	assert.Empty(t, withoutDetails.Schema)
	assert.Empty(t, withoutDetails.Fields)
	assert.Empty(t, withoutDetails.Examples)

	withoutChildren, err := Describe(DescribeInput{File: v1.GenerateFile, Path: "plugins", IncludeChildren: &falseValue})
	require.NoError(t, err)
	assert.NotContains(t, fieldPaths(withoutChildren.Fields), "plugins[].out")

	first, err := Describe(DescribeInput{File: v1.GenerateFile, Path: "plugins[].out"})
	require.NoError(t, err)
	first.Schema["type"] = "mutated"
	second, err := Describe(DescribeInput{File: v1.GenerateFile, Path: "plugins[].out"})
	require.NoError(t, err)
	assert.Equal(t, "string", second.Schema["type"])
}

func TestDescribeExamplesAreValidV1(t *testing.T) {
	t.Parallel()

	for _, file := range []string{v1.PolicyFile, v1.GenerateFile} {
		file := file
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			got, err := Describe(DescribeInput{File: file})
			require.NoError(t, err)
			require.NotEmpty(t, got.Examples)
			for _, example := range got.Examples {
				switch file {
				case v1.PolicyFile:
					assert.Empty(t, v1.ValidatePolicyYAML([]byte(example.YAML)), example.Title)
					_, err = v1.ParsePolicy(strings.NewReader(example.YAML))
				case v1.GenerateFile:
					assert.Empty(t, v1.ValidateGenerateYAML([]byte(example.YAML)), example.Title)
					_, err = v1.ParseGenerate(strings.NewReader(example.YAML))
				}
				require.NoError(t, err, example.Title)
			}
		})
	}
}

func TestRegisterToolViaMCP(t *testing.T) {
	t.Parallel()

	server := mcp.NewServer(&mcp.Implementation{Name: "easyp-test", Version: "v1"}, nil)
	RegisterTool(server)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(context.Background(), serverTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, serverSession.Close()) })

	client := mcp.NewClient(&mcp.Implementation{Name: "easyp-test-client", Version: "v1"}, nil)
	clientSession, err := client.Connect(context.Background(), clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, clientSession.Close()) })

	listed, err := clientSession.ListTools(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, listed.Tools, 1)
	assert.Equal(t, ToolName, listed.Tools[0].Name)

	response, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: ToolName, Arguments: map[string]any{"file": v1.GenerateFile, "path": "plugins[0].out"},
	})
	require.NoError(t, err)
	require.False(t, response.IsError)
	data, err := json.Marshal(response.StructuredContent)
	require.NoError(t, err)
	var described DescribeOutput
	require.NoError(t, json.NewDecoder(bytes.NewReader(data)).Decode(&described))
	assert.Equal(t, v1.GenerateFile, described.File)
	assert.Equal(t, "plugins[].out", described.SelectedPath)

	invalid, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: ToolName, Arguments: map[string]any{"file": v1.GenerateFile, "path": "generate.inputs"},
	})
	require.NoError(t, err)
	assert.True(t, invalid.IsError)
}

func fieldPaths(fields []FieldDoc) []string {
	paths := make([]string, 0, len(fields))
	for _, field := range fields {
		paths = append(paths, field.Path)
	}
	return paths
}
