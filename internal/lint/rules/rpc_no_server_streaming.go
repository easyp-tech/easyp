package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCNoServerStreaming)(nil)

// RPCNoServerStreaming this rule checks that RPCs aren't server streaming.
type RPCNoServerStreaming struct {
}

// Validate implements lint.Rule.
func (R RPCNoServerStreaming) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCResponse.IsStream {
				res = append(res, BuildError(rpc.Meta.Pos, rpc.RPCName, ErrRPCServerStreaming))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
