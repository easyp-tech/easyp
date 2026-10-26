package plugin

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

// marshalRequest encodes the request deterministically: custom options are
// dynamic messages whose field order is otherwise random between runs.
func marshalRequest(request *pluginpb.CodeGeneratorRequest) ([]byte, error) {
	return proto.MarshalOptions{Deterministic: true}.Marshal(request)
}
