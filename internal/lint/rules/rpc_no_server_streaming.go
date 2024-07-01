package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCNoServerStreaming)(nil)

// RPCNoServerStreaming this rule checks that RPCs aren't server streaming.
type RPCNoServerStreaming struct {
}

// Name implements lint.Rule.
func (r *RPCNoServerStreaming) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(r).Elem().Name())
}

// Validate implements lint.Rule.
func (r *RPCNoServerStreaming) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCResponse.IsStream {
				res = append(res, BuildError(protoInfo.Path, rpc.Meta.Pos, rpc.RPCName, lint.ErrRPCServerStreaming))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
