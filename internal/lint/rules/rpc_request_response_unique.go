package rules

import (
	"github.com/samber/lo"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCRequestResponseUnique)(nil)

// RPCRequestResponseUnique checks that RPCs request and response types are only used in one RPC.
type RPCRequestResponseUnique struct {
}

// Message implements lint.Rule.
func (r *RPCRequestResponseUnique) Message() string {
	return "request and response types must be unique across all RPCs"
}

// Validate implements lint.Rule.
func (r *RPCRequestResponseUnique) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	var messages []string

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if !lo.Contains(messages, rpc.RPCRequest.MessageType) {
				messages = append(messages, rpc.RPCRequest.MessageType)
			} else {
				res = append(res, lint.BuildError(rpc.Meta.Pos, rpc.RPCRequest.MessageType, r.Message()))
			}
			if !lo.Contains(messages, rpc.RPCResponse.MessageType) {
				messages = append(messages, rpc.RPCResponse.MessageType)
			} else {
				res = append(res, lint.BuildError(rpc.Meta.Pos, rpc.RPCResponse.MessageType, r.Message()))
			}
		}
	}

	return res, nil
}
