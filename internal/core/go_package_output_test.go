package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestGoPackageOutputPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		output  string
		content string
		want    string
	}{
		{name: "source relative collision", output: "v1/user.pb.go", want: "user/v1/user.pb.go"},
		{name: "grpc companion", output: "v1/user_grpc.pb.go", content: "// source: v1/user.proto\npackage userv1\n", want: "user/v1/user_grpc.pb.go"},
		{name: "already package relative", output: "order/v1/order.pb.go", want: "order/v1/order.pb.go"},
		{name: "non go output", output: "v1/user_pb2.py", want: "v1/user_pb2.py"},
		{name: "disabled option keeps effective directory", output: "api/models.pb.go", want: "api/models.pb.go"},
		{name: "marker directory follows effective import", output: "api/marker.pb.go", want: "api/marker/marker.pb.go"},
		{name: "longer proto basename", output: "v1/user_types.pb.go", want: "types/v1/user_types.pb.go"},
		{name: "import relative plugin output", output: "example.test/generated/user/v1/user.pb.go", want: "example.test/generated/user/v1/user.pb.go"},
		{name: "unknown source header", output: "v1/user.pb.go", content: "// source: dependency.proto\npackage dep\n", want: "v1/user.pb.go"},
		{name: "proto shares the grpc suffix", output: "v1/user_grpc.pb.go", content: "// source: v1/user_grpc.proto\npackage typesv1\n", want: "types/v1/user_grpc.pb.go"},
		{name: "body comments are not a source header", output: "v1/user.pb.go", content: "package userv1\n// source: dependency.proto\n", want: "user/v1/user.pb.go"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			directories := map[string]string{
				"v1/user.proto": "user/v1", "order/v1/order.proto": "order/v1",
				"api/models.proto": "api", "api/marker.proto": "api/marker",
				"v1/user_types.proto": "types/v1", "v1/user_grpc.proto": "types/v1",
			}
			file := &pluginpb.CodeGeneratorResponse_File{Name: proto.String(tt.output), Content: proto.String(tt.content)}

			assert.Equal(t, tt.want, goPackageOutputPath(file, directories))
		})
	}
}

func TestGoPackageOutputDirs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prefix   string
		packages map[string]string
		files    []string
		want     map[string]string
	}{
		{
			name: "effective package alias is not part of the directory", prefix: "example.test/generated",
			packages: map[string]string{"api/models.proto": "example.test/generated/custom;custom"},
			files:    []string{"api/models.proto"}, want: map[string]string{"api/models.proto": "custom"},
		},
		{
			name: "only plugin targets are mapped", prefix: "example.test/generated",
			packages: map[string]string{"api/models.proto": "example.test/generated/api", "dep.proto": "example.test/generated/dependency"},
			files:    []string{"api/models.proto"}, want: map[string]string{"api/models.proto": "api"},
		},
		{
			name: "external package keeps its plugin location", prefix: "example.test/generated",
			packages: map[string]string{"dep.proto": "example.test/other/dep"},
			files:    []string{"dep.proto"}, want: map[string]string{},
		},
		{
			name: "import prefix boundary", prefix: "example.test/gen",
			packages: map[string]string{"dep.proto": "example.test/generated/dep"},
			files:    []string{"dep.proto"}, want: map[string]string{},
		},
		{
			name: "SDK root package", prefix: "example.test/generated",
			packages: map[string]string{"root.proto": "example.test/generated;root"},
			files:    []string{"root.proto"}, want: map[string]string{"root.proto": "."},
		},
		{
			name: "marker uses stable parent", prefix: "example.test/generated/{{file_path}}",
			packages: map[string]string{"api/models.proto": "example.test/generated/api/models"},
			files:    []string{"api/models.proto"}, want: map[string]string{"api/models.proto": "api/models"},
		},
		{
			name: "marker can share a segment with literal text", prefix: "example.test/generated/prefix_{{file_dir}}",
			packages: map[string]string{"api/models.proto": "example.test/generated/prefix_api"},
			files:    []string{"api/models.proto"}, want: map[string]string{"api/models.proto": "prefix_api"},
		},
		{name: "empty prefix"},
		{name: "marker without a stable parent", prefix: "{{file_path}}"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var descriptors []*descriptorpb.FileDescriptorProto
			for file, goPackage := range tt.packages {
				descriptors = append(descriptors, &descriptorpb.FileDescriptorProto{
					Name: proto.String(file), Options: &descriptorpb.FileOptions{GoPackage: proto.String(goPackage)},
				})
			}

			directories := goPackageOutputDirs(tt.prefix, tt.files, descriptors)

			assert.Equal(t, tt.want, directories)
		})
	}
}
