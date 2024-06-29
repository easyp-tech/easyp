package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RpcPascalCase)(nil)

// RpcPascalCase this rule checks that RPCs are PascalCase.
type RpcPascalCase struct{}

// Validate implements lint.Rule.
func (c *RpcPascalCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+([A-Z][a-z]+)*$")
	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if !pascalCase.MatchString(rpc.RPCName) {
				res = AppendError(res, RPC_PASCAL_CASE, rpc.Meta.Pos, rpc.RPCName, rpc.Comments)
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
