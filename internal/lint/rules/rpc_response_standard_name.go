package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCResponseStandardName)(nil)

// RPCResponseStandardName checks that RPC response type names are RPCNameResponse or ServiceNameRPCNameResponse.
type RPCResponseStandardName struct {
}

// Validate implements lint.Rule.
func (R RPCResponseStandardName) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCResponse.MessageType != rpc.RPCName+"Response" && rpc.RPCResponse.MessageType != service.ServiceName+rpc.RPCName+"Response" {
				res = append(res, BuildError(rpc.Meta.Pos, rpc.RPCResponse.MessageType, ErrRPCResponseStandardName))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
