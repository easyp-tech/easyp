package plugin

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/easyp-tech/easyp/internal/compilerbackend"
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
	compiled, err := compilerbackend.Compile(t.Context(), compilerbackend.Request{
		Files: []string{"api.proto"},
		Accessor: compilerbackend.BytesAccessor(func(name string) ([]byte, error) {
			src, ok := sources[name]
			if !ok {
				return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
			}
			return []byte(src), nil
		}),
		StandardImports: true,
	})
	require.NoError(t, err)
	request := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"api.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{compiled.Protos["api.proto"]},
	}

	want, err := marshalRequest(request)
	require.NoError(t, err)
	for range 100 {
		got, err := marshalRequest(request)
		require.NoError(t, err)
		require.Equal(t, want, got)
	}
}
