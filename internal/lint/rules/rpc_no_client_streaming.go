package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCNoClientStreaming)(nil)

// RPCNoClientStreaming this rule checks that RPCs aren't client streaming.
type RPCNoClientStreaming struct {
}

// Name implements lint.Rule.
func (r *RPCNoClientStreaming) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(r).Elem().Name())
}

// Validate implements lint.Rule.
func (r *RPCNoClientStreaming) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCRequest.IsStream {
				res = append(res, BuildError(protoInfo.Path, rpc.Meta.Pos, rpc.RPCName, lint.ErrRPCClientStreaming))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
