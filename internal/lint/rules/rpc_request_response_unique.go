package rules

import (
	"github.com/samber/lo"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCRequestResponseUnique)(nil)

// RPCRequestResponseUnique checks that RPCs request and response types are only used in one RPC.
type RPCRequestResponseUnique struct {
}

// Validate implements lint.Rule.
func (R RPCRequestResponseUnique) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	var messages []string

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if !lo.Contains(messages, rpc.RPCRequest.MessageType) {
				messages = append(messages, rpc.RPCRequest.MessageType)
			} else {
				res = append(res, BuildError(rpc.Meta.Pos, rpc.RPCRequest.MessageType, ErrRPCRequestResponseUnique))
			}
			if !lo.Contains(messages, rpc.RPCResponse.MessageType) {
				messages = append(messages, rpc.RPCResponse.MessageType)
			} else {
				res = append(res, BuildError(rpc.Meta.Pos, rpc.RPCResponse.MessageType, ErrRPCRequestResponseUnique))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
