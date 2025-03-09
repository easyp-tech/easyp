package rules

import (
	"github.com/samber/lo"

	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*RPCRequestResponseUnique)(nil)

// RPCRequestResponseUnique checks that RPCs request and response types are only used in one RPC.
type RPCRequestResponseUnique struct {
}

// Message implements lint.Rule.
func (r *RPCRequestResponseUnique) Message() string {
	return "request and response types must be unique across all RPCs"
}

// Validate implements lint.Rule.
func (r *RPCRequestResponseUnique) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	var messages []string

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if !lo.Contains(messages, rpc.RPCRequest.MessageType) {
				messages = append(messages, rpc.RPCRequest.MessageType)
			} else {
				res = core.AppendIssue(res, r, rpc.Meta.Pos, rpc.RPCRequest.MessageType, rpc.Comments)
			}
			if !lo.Contains(messages, rpc.RPCResponse.MessageType) {
				messages = append(messages, rpc.RPCResponse.MessageType)
			} else {
				res = core.AppendIssue(res, r, rpc.Meta.Pos, rpc.RPCResponse.MessageType, rpc.Comments)
			}
		}
	}

	return res, nil
}
