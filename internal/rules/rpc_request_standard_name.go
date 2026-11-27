package rules

import (
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*RPCRequestStandardName)(nil)

// RPCRequestStandardName this rule checks that RPC request types are named <RPC>Request or <Service><RPC>Request.
type RPCRequestStandardName struct{}

// Message implements lint.Rule.
func (r *RPCRequestStandardName) Message() string {
	return "rpc request should have suffix 'Request'"
}

// Validate implements lint.Rule.
func (r *RPCRequestStandardName) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{RPC: func(service *protoast.Service, rpc *protoast.RPC) {
		name := rpc.Request.Type[strings.LastIndex(rpc.Request.Type, ".")+1:]
		expected := toPascalCase(rpc.Name) + "Request"
		if name != expected && name != toPascalCase(service.Name)+expected {
			res = core.AppendIssue(res, r, rpc.Request.Pos, rpc.Request.Type)
		}
	}})

	return res, nil
}
