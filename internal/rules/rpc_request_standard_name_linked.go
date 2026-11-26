package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*RPCRequestStandardName)(nil)

// RPCRequestStandardName (research inventory #29, main-diverged; threat
// T-04-14): the WRITTEN request type string must equal RPCName+"Request" or
// ServiceName+RPCName+"Request" EXACTLY — a qualified written type
// ("pkg.v1.GetFooRequest") fails both alternatives, and there is no pascal
// normalization and no dot-segment splitting (the lab's last-dot-segment
// compare would accept the qualified form — not ported). The position is
// the rpc keyword (main Meta.Pos ≡ Decl.Pos), never the lab's type-token
// position.
type RPCRequestStandardName struct{}

// Message implements lint.Rule.
func (r *RPCRequestStandardName) Message() string {
	return "rpc request should have suffix 'Request'"
}

// Validate implements lint.Rule.
func (r *RPCRequestStandardName) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range info.File.Services {
		for _, rpc := range service.RPCs {
			written := rpc.Request.Type
			if written != rpc.Name+"Request" && written != service.Name+rpc.Name+"Request" {
				res = core.AppendIssue(res, r, rpc.Pos, written)
			}
		}
	}

	return res, nil
}
