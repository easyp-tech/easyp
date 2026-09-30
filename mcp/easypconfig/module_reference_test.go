package easypconfig

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	yamlvalidator "github.com/Yakwilik/go-yamlvalidator"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestModuleReferenceExamplesAndGrammar(t *testing.T) {
	t.Parallel()
	out, err := Describe(DescribeInput{File: v1.ModuleFile})
	require.NoError(t, err)
	assert.Equal(t, "text", out.Format)
	assert.Nil(t, out.Schema)
	require.NotNil(t, out.Grammar)
	assert.Equal(t, "protobuf.mod/v1", out.Grammar.Dialect)
	assert.Contains(t, out.Grammar.Syntax, "module <identity>")
	assert.Contains(t, fieldPaths(out.Fields), "replace[].target")
	require.Len(t, out.Examples, 4)
	for _, example := range out.Examples {
		assert.Equal(t, "text", example.Format)
		assert.Empty(t, example.YAML)
		_, err := v1.ParseModule(strings.NewReader(example.Text))
		require.NoError(t, err, example.Title)
		data, err := json.Marshal(example)
		require.NoError(t, err)
		assert.NotContains(t, string(data), `"yaml"`)
	}
	_, err = SchemaByPathFor(v1.ModuleFile)
	require.ErrorContains(t, err, "text manifest")
}

func TestLockReferenceUsesActualSchemaAndValidExamples(t *testing.T) {
	t.Parallel()
	out, err := Describe(DescribeInput{File: v1.LockFile})
	require.NoError(t, err)
	assert.Equal(t, "yaml", out.Format)
	assert.Nil(t, out.Grammar)
	raw, err := v1.SchemaJSON("protobuf.lock")
	require.NoError(t, err)
	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))
	assert.Equal(t, schema, out.Schema)
	for _, example := range out.Examples {
		_, err := v1.ParseLock(strings.NewReader(example.YAML))
		require.NoError(t, err, example.Title)
		assert.Empty(t, example.Text)
		compiled, err := yamlvalidator.CompileJSONSchema(raw)
		require.NoError(t, err)
		checked := yamlvalidator.NewValidator(compiled).ValidateWithOptions([]byte(example.YAML), yamlvalidator.ValidationContext{StrictKeys: true, StrictTypes: true})
		assert.Empty(t, checked.Collector.All(), example.Title)
	}
	field, err := Describe(DescribeInput{File: v1.LockFile, Path: "$.modules[2].commit"})
	require.NoError(t, err)
	require.Len(t, field.Fields, 1)
	assert.True(t, field.Fields[0].Required)
	assert.Equal(t, "modules[].commit", field.SelectedPath)
	assert.Equal(t, "string", field.Fields[0].Type)
	version, err := Describe(DescribeInput{File: v1.LockFile, Path: "version"})
	require.NoError(t, err)
	assert.Equal(t, "integer", version.Fields[0].Type)
	assert.Equal(t, float64(1), version.Schema["const"])
}

func TestModuleReferenceSelectionAndFlags(t *testing.T) {
	t.Parallel()
	disabled := false
	for _, tc := range []struct{ path, want string }{
		{path: "", want: "$"}, {path: "root", want: "$"}, {path: "$.require[12].version", want: "require[].version"}, {path: ".replace[*].target", want: "replace[].target"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			out, err := Describe(DescribeInput{File: v1.ModuleFile, Path: tc.path})
			require.NoError(t, err)
			assert.Equal(t, tc.want, out.SelectedPath)
			require.NotEmpty(t, out.Grammar.Syntax)
		})
	}
	only, err := Describe(DescribeInput{File: v1.ModuleFile, Path: "require", IncludeChildren: &disabled})
	require.NoError(t, err)
	require.Len(t, only.Fields, 1)
	assert.Equal(t, "require", only.Fields[0].Path)
	omitted, err := Describe(DescribeInput{File: v1.ModuleFile, IncludeChildren: &disabled, IncludeSchema: &disabled, IncludeFields: &disabled, IncludeExamples: &disabled})
	require.NoError(t, err)
	assert.Nil(t, omitted.Grammar)
	assert.Nil(t, omitted.Schema)
	assert.Empty(t, omitted.Fields)
	assert.Empty(t, omitted.Examples)
	limit := 1
	limited, err := Describe(DescribeInput{File: v1.ModuleFile, ExamplesLimit: &limit})
	require.NoError(t, err)
	assert.Len(t, limited.Examples, 1)
}

func TestReferenceRejectsInvalidSelectorsAndLimits(t *testing.T) {
	t.Parallel()
	for _, file := range []string{v1.PolicyFile, v1.GenerateFile, v1.ModuleFile, v1.LockFile} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			for _, limit := range []int{-1, 0, 51} {
				_, err := Describe(DescribeInput{File: file, ExamplesLimit: &limit})
				require.ErrorContains(t, err, "examples_limit")
			}
			_, err := Describe(DescribeInput{File: file, Path: "unknown.directive"})
			require.ErrorContains(t, err, "unknown path")
		})
	}
	for _, file := range []string{"../protobuf.lock", "https://example.test/protobuf.mod", "${HOME}/protobuf.mod"} {
		_, err := Describe(DescribeInput{File: file})
		require.ErrorContains(t, err, "unknown config file")
	}
}

func TestReferenceResultsDoNotShareState(t *testing.T) {
	t.Parallel()
	for _, file := range []string{v1.ModuleFile, v1.LockFile, v1.PolicyFile, v1.GenerateFile} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			before, err := Describe(DescribeInput{File: file})
			require.NoError(t, err)
			changed, err := Describe(DescribeInput{File: file})
			require.NoError(t, err)
			for i := range changed.Fields {
				changed.Fields[i].Description = "changed"
			}
			for i := range changed.Examples {
				changed.Examples[i].Text = "changed"
				changed.Examples[i].YAML = "changed"
				for j := range changed.Examples[i].Paths {
					changed.Examples[i].Paths[j] = "changed"
				}
			}
			for i := range changed.Notes {
				changed.Notes[i] = "changed"
			}
			if changed.Grammar != nil {
				changed.Grammar.Syntax[0] = "changed"
				changed.Grammar.Notes[0] = "changed"
			}
			if changed.Schema != nil {
				changed.Schema["properties"].(map[string]any)["new"] = true
			}
			again, err := Describe(DescribeInput{File: file})
			require.NoError(t, err)
			assert.Equal(t, before, again)
		})
	}
}

func TestMCPInputSchemaMatchesTypedArguments(t *testing.T) {
	t.Parallel()
	properties := describeInputSchema()["properties"].(map[string]any)
	typ := reflect.TypeFor[DescribeInput]()
	assert.Len(t, properties, typ.NumField())
	for i := range typ.NumField() {
		name := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
		assert.Contains(t, properties, name)
	}
}

func TestModuleReferenceActualMCPRegistration(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	server := mcp.NewServer(&mcp.Implementation{Name: "reference-test", Version: "v1"}, nil)
	RegisterTool(server)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { require.NoError(t, ss.Close()) }()
	client := mcp.NewClient(&mcp.Implementation{Name: "reference-client", Version: "v1"}, nil)
	cs, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { require.NoError(t, cs.Close()) }()
	listed, err := cs.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, listed.Tools, 1)
	raw, err := json.Marshal(listed.Tools[0].InputSchema)
	require.NoError(t, err)
	for _, file := range []string{v1.PolicyFile, v1.GenerateFile, v1.ModuleFile, v1.LockFile} {
		assert.Contains(t, string(raw), file)
	}
	for _, file := range []string{v1.ModuleFile, v1.LockFile} {
		out, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: ToolName, Arguments: map[string]any{"file": file}})
		require.NoError(t, err)
		require.False(t, out.IsError, "%+v", out.Content)
		require.NotNil(t, out.StructuredContent)
	}
	for _, args := range []map[string]any{{"file": "/tmp/protobuf.mod"}, {"file": v1.ModuleFile, "path": "module.unknown"}, {"examples_limit": 0}, {"examples_limit": 51}, {"unexpected": true}} {
		out, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: ToolName, Arguments: args})
		require.NoError(t, err)
		assert.True(t, out.IsError)
	}
}
