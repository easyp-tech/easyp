package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCNoClientStreaming)(nil)

// RPCNoClientStreaming this rule checks that RPCs aren't client streaming.
type RPCNoClientStreaming struct {
}

// Validate implements lint.Rule.
func (R RPCNoClientStreaming) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCRequest.IsStream {
				res = AppendError(res, RPC_NO_CLIENT_STREAMING, rpc.Meta.Pos, rpc.RPCName, rpc.Comments)
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
