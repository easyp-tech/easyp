package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestGoPackageOutputPath(t *testing.T) {
	t.Parallel()

	descriptors := []*descriptorpb.FileDescriptorProto{
		{Name: proto.String("v1/user.proto"), Package: proto.String("user.v1")},
		{Name: proto.String("order/v1/order.proto"), Package: proto.String("order.v1")},
	}
	for _, tt := range []struct {
		name   string
		output string
		want   string
	}{
		{name: "source relative collision", output: "v1/user.pb.go", want: "user/v1/user.pb.go"},
		{name: "grpc companion", output: "v1/user_grpc.pb.go", want: "user/v1/user_grpc.pb.go"},
		{name: "already package relative", output: "order/v1/order.pb.go", want: "order/v1/order.pb.go"},
		{name: "non go output", output: "v1/user_pb2.py", want: "v1/user_pb2.py"},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, goPackageOutputPath(tt.output, []string{"v1/user.proto", "order/v1/order.proto"}, descriptors))
		})
	}
}
