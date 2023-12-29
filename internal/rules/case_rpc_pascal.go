package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*RpcPascalCase)(nil)

// RpcPascalCase is a rule for checking name of rpc for pascal case.
type RpcPascalCase struct{}

// Validate implements Rule.
func (c *RpcPascalCase) Validate(protoInfo core.ProtoInfo) []error {
	var res []error
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+([A-Z][a-z]+)*$")
	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if !pascalCase.MatchString(rpc.RPCName) {
				res = append(res, buildError(rpc.Meta.Pos, rpc.RPCName, core.ErrRpcPascalCase))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
