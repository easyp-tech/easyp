package easypconfig

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/easyp-tech/easyp/internal/rules"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestDescribe_GitRepoPath_NoOutField(t *testing.T) {
	t.Parallel()

	out, err := Describe(DescribeInput{Path: "generate.inputs[*].git_repo"})
	require.NoError(t, err)

	require.Equal(t, SchemaVersion, out.SchemaVersion)
	require.Equal(t, "generate.inputs[].git_repo", out.SelectedPath)
	require.Contains(t, fieldPaths(out.Fields), "generate.inputs[].git_repo.url")
	require.NotContains(t, fieldPaths(out.Fields), "generate.inputs[].git_repo.out")

	props, ok := nestedMap(out.Schema, "properties")
	require.True(t, ok, "expected properties in schema fragment")
	_, hasOut := props["out"]
	require.False(t, hasOut, "git_repo.out must not exist in schema")
}

func TestDescribe_UnknownPath(t *testing.T) {
	t.Parallel()

	_, err := Describe(DescribeInput{Path: "unknown.section"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown path")
}

func TestDescribe_FlagsAndExamplesLimitClamp(t *testing.T) {
	t.Parallel()

	f := false
	out, err := Describe(DescribeInput{
		Path:            "$",
		IncludeSchema:   &f,
		IncludeFields:   &f,
		IncludeExamples: &f,
	})
	require.NoError(t, err)
	require.Nil(t, out.Schema)
	require.Empty(t, out.Fields)
	require.Empty(t, out.Examples)

	tv := true
	zero := 0
	out, err = Describe(DescribeInput{
		Path:            "$",
		IncludeExamples: &tv,
		ExamplesLimit:   &zero,
	})
	require.NoError(t, err)
	require.Len(t, out.Examples, 1)
}

func TestDescribe_IncludeChildrenToggle(t *testing.T) {
	t.Parallel()

	f := false
	withoutChildren, err := Describe(DescribeInput{
		Path:            "generate",
		IncludeChildren: &f,
	})
	require.NoError(t, err)
	require.NotContains(t, fieldPaths(withoutChildren.Fields), "generate.inputs[].directory.path")

	withChildren, err := Describe(DescribeInput{
		Path: "generate",
	})
	require.NoError(t, err)
	require.Contains(t, fieldPaths(withChildren.Fields), "generate.inputs[].directory.path")
}

func TestDescribe_PathNormalization_GenericArrayIndex(t *testing.T) {
	t.Parallel()

	out, err := Describe(DescribeInput{Path: "$.generate.inputs[12].git_repo"})
	require.NoError(t, err)
	require.Equal(t, "generate.inputs[].git_repo", out.SelectedPath)

	out, err = Describe(DescribeInput{Path: "generate.plugins[99]"})
	require.NoError(t, err)
	require.Equal(t, "generate.plugins[]", out.SelectedPath)
}

func TestDescribe_LintAllowedValuesFromRules(t *testing.T) {
	t.Parallel()

	out, err := Describe(DescribeInput{Path: "lint"})
	require.NoError(t, err)

	var lintUse FieldDoc
	var found bool
	for _, field := range out.Fields {
		if field.Path == "lint.use" {
			lintUse = field
			found = true
			break
		}
	}
	require.True(t, found, "lint.use field must be present")

	require.ElementsMatch(t, rules.AllLintUseValues(), lintUse.AllowedValues)
}

func TestDescribe_OutputSlicesAreIndependent(t *testing.T) {
	t.Parallel()

	out, err := Describe(DescribeInput{Path: "lint"})
	require.NoError(t, err)
	require.NotEmpty(t, out.Examples)

	lintFieldIdx := -1
	for i := range out.Fields {
		if out.Fields[i].Path == "lint.use" {
			lintFieldIdx = i
			break
		}
	}
	require.NotEqual(t, -1, lintFieldIdx)
	require.NotEmpty(t, out.Fields[lintFieldIdx].AllowedValues)
	require.NotEmpty(t, out.Examples[0].Paths)

	firstAllowed := out.Fields[lintFieldIdx].AllowedValues[0]
	firstExampleTitle := out.Examples[0].Title
	firstExamplePath := out.Examples[0].Paths[0]

	out.Fields[lintFieldIdx].AllowedValues[0] = "__MUTATED__"
	out.Examples[0].Paths[0] = "__MUTATED__"

	outAfterMutation, err := Describe(DescribeInput{Path: "lint"})
	require.NoError(t, err)

	var lintUseAfter FieldDoc
	found := false
	for _, field := range outAfterMutation.Fields {
		if field.Path == "lint.use" {
			lintUseAfter = field
			found = true
			break
		}
	}
	require.True(t, found)
	require.Equal(t, firstAllowed, lintUseAfter.AllowedValues[0])
	require.NotEqual(t, "__MUTATED__", lintUseAfter.AllowedValues[0])

	var exampleAfter Example
	found = false
	for _, example := range outAfterMutation.Examples {
		if example.Title == firstExampleTitle {
			exampleAfter = example
			found = true
			break
		}
	}
	require.True(t, found)
	require.Equal(t, firstExamplePath, exampleAfter.Paths[0])
	require.NotEqual(t, "__MUTATED__", exampleAfter.Paths[0])
}

func TestCollectExamples_DeduplicatesByFullExampleContent(t *testing.T) {
	t.Parallel()

	s := spec{
		DocsByPath: map[string]nodeDoc{
			"a": {
				Examples: []Example{
					{Title: "same_title", YAML: "x: 1", Paths: []string{"a"}},
					{Title: "same_title", YAML: "x: 2", Paths: []string{"a"}},
					{Title: "same_title", YAML: "x: 1", Paths: []string{"a"}},
				},
			},
		},
	}

	got := s.collectExamples([]string{"a"}, 10)
	require.Len(t, got, 2)
	require.Equal(t, "x: 1", got[0].YAML)
	require.Equal(t, "x: 2", got[1].YAML)
}

func TestSchemaByPath_GitRepoOutAbsent(t *testing.T) {
	t.Parallel()

	index := SchemaByPath()
	node, ok := index["generate.inputs[].git_repo"]
	require.True(t, ok, "expected schema path generate.inputs[].git_repo")

	props, ok := nestedMap(node, "properties")
	require.True(t, ok)
	_, hasOut := props["out"]
	require.False(t, hasOut)
}

func TestMarshalConfigJSONSchema_Golden(t *testing.T) {
	t.Parallel()

	got, err := MarshalConfigJSONSchema()
	require.NoError(t, err)

	goldenPath := filepath.Join("..", "..", "schemas", "easyp-config-v1.schema.json")
	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err)

	var gotObj any
	var wantObj any
	require.NoError(t, json.Unmarshal(got, &gotObj))
	require.NoError(t, json.Unmarshal(want, &wantObj))
	require.Equal(t, wantObj, gotObj)
}

func TestRegisterTool(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	srv := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "v1.0.0"}, nil)
	RegisterTool(srv)

	mux := http.NewServeMux()
	mux.Handle("/mcp", mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{}))

	httpSrv := httptest.NewServer(mux)
	defer httpSrv.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: httpSrv.URL + "/mcp"}, nil)
	require.NoError(t, err)
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Contains(t, toolNames(tools.Tools), ToolName)

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: ToolName,
		Arguments: map[string]any{
			"path": "generate.plugins[]",
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)

	var out DescribeOutput
	decodeStructured(t, res, &out)
	require.Equal(t, "generate.plugins[]", out.SelectedPath)

	errRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: ToolName,
		Arguments: map[string]any{
			"path": "unknown.section",
		},
	})
	require.NoError(t, err)
	require.True(t, errRes.IsError)
	require.Contains(t, toolText(errRes), "unknown path")
}

func nestedMap(schema map[string]any, key string) (map[string]any, bool) {
	v, ok := schema[key]
	if !ok {
		return nil, false
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, false
	}
	return m, true
}

func fieldPaths(fields []FieldDoc) []string {
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		out = append(out, f.Path)
	}
	return out
}

func toolNames(tools []*mcp.Tool) []string {
	out := make([]string, 0, len(tools))
	for _, tool := range tools {
		out = append(out, tool.Name)
	}
	return out
}

func decodeStructured(t *testing.T, res *mcp.CallToolResult, dst any) {
	t.Helper()

	data, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, dst))
}

func toolText(res *mcp.CallToolResult) string {
	parts := make([]string, 0, len(res.Content))
	for _, c := range res.Content {
		if text, ok := c.(*mcp.TextContent); ok {
			parts = append(parts, text.Text)
		}
	}
	return strings.Join(parts, "\n")
}
