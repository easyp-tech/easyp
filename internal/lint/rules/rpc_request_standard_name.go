package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCRequestStandardName)(nil)

// RPCRequestStandardName checks that RPC request type names are RPCNameRequest or ServiceNameRPCNameRequest.
type RPCRequestStandardName struct {
}

// Name implements lint.Rule.
func (r *RPCRequestStandardName) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(r).Elem().Name())
}

// Validate implements lint.Rule.
func (r *RPCRequestStandardName) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCRequest.MessageType != rpc.RPCName+"Request" && rpc.RPCRequest.MessageType != service.ServiceName+rpc.RPCName+"Request" {
				res = append(res, BuildError(protoInfo.Path, rpc.Meta.Pos, rpc.RPCRequest.MessageType, lint.ErrRPCRequestStandardName))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
