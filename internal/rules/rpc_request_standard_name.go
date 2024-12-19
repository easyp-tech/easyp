package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*RPCRequestStandardName)(nil)

// RPCRequestStandardName checks that RPC request type names are RPCNameRequest or ServiceNameRPCNameRequest.
type RPCRequestStandardName struct {
}

// Message implements lint.Rule.
func (r *RPCRequestStandardName) Message() string {
	return "rpc request should have suffix 'Request'"
}

// Validate implements lint.Rule.
func (r *RPCRequestStandardName) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCRequest.MessageType != rpc.RPCName+"Request" && rpc.RPCRequest.MessageType != service.ServiceName+rpc.RPCName+"Request" {
				res = core.AppendIssue(res, r, rpc.Meta.Pos, rpc.RPCRequest.MessageType, rpc.Comments)
			}
		}
	}

	return res, nil
}
