package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCResponseStandardName)(nil)

// RPCResponseStandardName checks that RPC response type names are RPCNameResponse or ServiceNameRPCNameResponse.
type RPCResponseStandardName struct {
}

// Message implements lint.Rule.
func (r *RPCResponseStandardName) Message() string {
	return "rpc response should have suffix 'Response'"
}

// Validate implements lint.Rule.
func (r *RPCResponseStandardName) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCResponse.MessageType != rpc.RPCName+"Response" && rpc.RPCResponse.MessageType != service.ServiceName+rpc.RPCName+"Response" {
				res = append(res, lint.BuildError(r, rpc.Meta.Pos, rpc.RPCResponse.MessageType))
			}
		}
	}

	return res, nil
}
