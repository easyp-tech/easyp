package protoast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/easypc/protoc/syntax"
)

const sample = `// File comment.
syntax = "proto3";

package a.b.v1; // trailing package

import public "x/y.proto";

// Detached.

// Leading 1.
// Leading 2.
message Outer { // after brace
  repeated .a.b.v1.Inner inner = 1 [json_name = "i", (ext.opt).sub = "x" "y"];
  map<string, int32> counts = 2;
  oneof choice {
    string name = 3;
  }
  message Inner {}
}

enum Kind {
	KIND_UNSPECIFIED = 0;
	KIND_NEG = -1 [deprecated = true];
}

service Svc {
  rpc Get(stream Outer) returns (.a.b.v1.Outer) {
    option (google.api.http) = { get: "/v1" };
  }
}
// End of file.
`

func TestParse(t *testing.T) {
	f, err := Parse("a/b/v1/x.proto", []byte(sample))
	require.NoError(t, err)

	assert.Equal(t, "proto3", f.Syntax)
	require.Len(t, f.Packages, 1)
	assert.Equal(t, "a.b.v1", f.Packages[0].Name)
	assert.Equal(t, Position{Offset: 37, Line: 4, Column: 1}, f.Packages[0].Pos)
	require.Len(t, f.Packages[0].Trailing, 1)
	assert.Equal(t, "// trailing package", f.Packages[0].Trailing[0].Raw)

	require.Len(t, f.Imports, 1)
	assert.Equal(t, `"x/y.proto"`, f.Imports[0].Location)
	assert.Equal(t, "x/y.proto", f.Imports[0].Path)
	assert.True(t, f.Imports[0].Public)

	require.Len(t, f.Messages, 1)
	outer := f.Messages[0]
	assert.Equal(t, "Outer", outer.Name)
	require.Len(t, outer.Detached, 1)
	assert.Equal(t, "// Detached.", outer.Detached[0].Raw)
	require.Len(t, outer.Leading, 2)
	require.Len(t, outer.Trailing, 1)
	assert.Equal(t, "// after brace", outer.Trailing[0].Raw)

	require.Len(t, outer.Fields, 1)
	inner := outer.Fields[0]
	assert.Equal(t, "repeated", inner.Label)
	assert.Equal(t, ".a.b.v1.Inner", inner.Type)
	assert.Equal(t, "inner", inner.Name)
	assert.Equal(t, 3, inner.Pos.Column)
	require.Len(t, inner.Options, 2)
	assert.Equal(t, Option{Decl: inner.Options[0].Decl, Name: "json_name", Value: `"i"`}, *inner.Options[0])
	assert.Equal(t, "(ext.opt).sub", inner.Options[1].Name)
	assert.Equal(t, `"xy"`, inner.Options[1].Value)

	require.Len(t, outer.Maps, 1)
	assert.Equal(t, "string", outer.Maps[0].KeyType)
	assert.Equal(t, "int32", outer.Maps[0].Type)
	require.Len(t, outer.Oneofs, 1)
	require.Len(t, outer.Oneofs[0].Fields, 1)
	assert.Equal(t, "name", outer.Oneofs[0].Fields[0].Name)
	require.Len(t, outer.Messages, 1)

	require.Len(t, f.Enums, 1)
	values := f.Enums[0].Values
	require.Len(t, values, 2)
	assert.Equal(t, "-1", values[1].Number)
	assert.Equal(t, 9, values[1].Pos.Column, "a tab advances to the next multiple of 8, as in protoc")
	assert.Equal(t, values[1].Pos, values[1].NamePos)
	assert.Equal(t, 20, values[1].NumberPos.Column)
	assert.Equal(t, 9, outer.NamePos.Column)
	assert.Equal(t, 34, f.Services[0].RPCs[0].Response.Pos.Column)
	assert.Equal(t, 18, f.Services[0].RPCs[0].Request.Pos.Column, "after the stream keyword")
	require.Len(t, values[1].Options, 1)
	assert.Equal(t, "true", values[1].Options[0].Value)

	require.Len(t, f.Services, 1)
	rpc := f.Services[0].RPCs[0]
	assert.Equal(t, "Outer", rpc.Request.Type)
	assert.True(t, rpc.Request.Stream)
	assert.Equal(t, ".a.b.v1.Outer", rpc.Response.Type)
	assert.False(t, rpc.Response.Stream)
	require.Len(t, rpc.Options, 1)
	assert.Equal(t, "(google.api.http)", rpc.Options[0].Name)

	require.NotEmpty(t, f.Comments)
	assert.Same(t, &outer.Decl, f.Owner(outer.Detached[0]))
	last := f.Comments[len(f.Comments)-1]
	assert.Equal(t, "// End of file.", last.Raw)
	assert.Nil(t, f.Owner(last))
}

func TestParseSyntaxError(t *testing.T) {
	_, err := Parse("bad.proto", []byte(`syntax = "proto3"; message {`))
	require.Error(t, err)
}

func TestParseIgnoresSemanticDiagnostics(t *testing.T) {
	f, err := Parse("dup.proto", []byte(`syntax = "proto3"; message M { int32 a = 1; int32 b = 1; }`))
	require.NoError(t, err)
	require.Len(t, f.Messages[0].Fields, 2)
}

func TestOutline(t *testing.T) {
	f, err := Parse("x.proto", []byte(sample))
	require.NoError(t, err)
	out := f.Outline()
	assert.Equal(t, "a.b.v1", out.PackageName())
	require.Len(t, out.Messages, 1)
	assert.Equal(t, "Outer", out.Messages[0].Name)
	assert.Empty(t, out.Messages[0].Fields)
	assert.Empty(t, out.Comments)
	assert.NotSame(t, f.Packages[0], out.Packages[0])
	assert.Len(t, f.Messages[0].Fields, 1, "outline must not change the source file")
}

func TestOutlineOfMatchesOutline(t *testing.T) {
	src := sample + "extend Outer { string extra = 100; }\n"
	f, err := syntax.Parse("x.proto", []byte(src))
	require.NoError(t, err)
	want, got := Build(f).Outline(), OutlineOf(f)
	assert.Equal(t, want.PackageName(), got.PackageName())
	require.Len(t, got.Messages, len(want.Messages))
	require.Len(t, got.Enums, len(want.Enums))
	require.Len(t, got.Extends, 1)
	assert.Equal(t, "extra", got.Extends[0].Fields[0].Name)
	assert.Equal(t, want.Messages[0].Name, got.Messages[0].Name)
}
