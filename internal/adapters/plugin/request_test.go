package plugin

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/easyp-tech/easyp/internal/compiler"
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
	resolve := func(name string) (io.ReadCloser, error) {
		source, ok := sources[name]
		if !ok {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
		}
		return io.NopCloser(strings.NewReader(source)), nil
	}
	files, err := compiler.Compile(t.Context(), []string{"api.proto"}, compiler.Options{
		Resolve: compiler.StandardImports(resolve),
	})
	require.NoError(t, err)
	request := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"api.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{protodesc.ToFileDescriptorProto(files[0])},
	}

	want, err := marshalRequest(request)
	require.NoError(t, err)
	for range 100 {
		got, err := marshalRequest(request)
		require.NoError(t, err)
		require.Equal(t, want, got)
	}
}
