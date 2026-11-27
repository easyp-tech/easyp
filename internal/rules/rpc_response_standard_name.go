package rules

import (
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*RPCResponseStandardName)(nil)

// RPCResponseStandardName this rule checks that RPC response types are named <RPC>Response or <Service><RPC>Response.
type RPCResponseStandardName struct{}

// Message implements lint.Rule.
func (r *RPCResponseStandardName) Message() string {
	return "rpc response should have suffix 'Response'"
}

// Validate implements lint.Rule.
func (r *RPCResponseStandardName) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{RPC: func(service *protoast.Service, rpc *protoast.RPC) {
		name := rpc.Response.Type[strings.LastIndex(rpc.Response.Type, ".")+1:]
		expected := toPascalCase(rpc.Name) + "Response"
		if name != expected && name != toPascalCase(service.Name)+expected {
			res = core.AppendIssue(res, r, rpc.Response.Pos, rpc.Response.Type)
		}
	}})

	return res, nil
}
