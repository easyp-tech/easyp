package plugin

import (
	"testing"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/protoutil"
	"github.com/bufbuild/protocompile/wellknownimports"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestMarshalRequestIsDeterministic(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"options.proto": `syntax = "proto3";
package options;
import "google/protobuf/descriptor.proto";
message Meta { string a = 1; string b = 2; int32 c = 3; bool d = 4; repeated string e = 5; }
extend google.protobuf.MessageOptions { Meta meta = 50001; string tag = 50002; }
`,
		"api.proto": `syntax = "proto3";
package api;
import "options.proto";
message Item {
  option (options.meta) = { a: "a" b: "b" c: 3 d: true e: ["x", "y"] };
  option (options.tag) = "item";
  string id = 1;
}
`,
	}
	compiler := protocompile.Compiler{Resolver: wellknownimports.WithStandardImports(&protocompile.SourceResolver{
		Accessor: protocompile.SourceAccessorFromMap(sources),
	})}
	files, err := compiler.Compile(t.Context(), "api.proto")
	require.NoError(t, err)
	request := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"api.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{protoutil.ProtoFromFileDescriptor(files[0])},
	}

	want, err := marshalRequest(request)
	require.NoError(t, err)
	for range 100 {
		got, err := marshalRequest(request)
		require.NoError(t, err)
		require.Equal(t, want, got)
	}
}
