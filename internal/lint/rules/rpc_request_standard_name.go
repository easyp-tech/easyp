package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCRequestStandardName)(nil)

// RPCRequestStandardName checks that RPC request type names are RPCNameRequest or ServiceNameRPCNameRequest.
type RPCRequestStandardName struct {
}

// Message implements lint.Rule.
func (r *RPCRequestStandardName) Message() string {
	return "rpc request should have suffix 'Request'"
}

// Validate implements lint.Rule.
func (r *RPCRequestStandardName) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCRequest.MessageType != rpc.RPCName+"Request" && rpc.RPCRequest.MessageType != service.ServiceName+rpc.RPCName+"Request" {
				res = append(res, lint.BuildError(r, rpc.Meta.Pos, rpc.RPCRequest.MessageType))
			}
		}
	}

	return res, nil
}
