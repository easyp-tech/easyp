package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*RPCNoServerStreamingLinked)(nil)

// RPCNoServerStreamingLinked is the linked-engine port of RPCNoServerStreaming
// (research inventory #41, verbatim-from-lab): the etalon's streaming
// predicate (rpc.Response.Stream — the lab's IsStream field renamed by main's
// protoast model) over main's shallow services→rpcs walk; the finding
// position is the declaration keyword (rpc.Meta.Pos ≡ Decl.Pos — main and
// etalon carried tests assert the identical literal on the shared auth
// fixture, the alias-only test diff).
type RPCNoServerStreamingLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (r *RPCNoServerStreamingLinked) LintRuleName() string {
	return "RPC_NO_SERVER_STREAMING"
}

// Message implements lint.Rule.
func (r *RPCNoServerStreamingLinked) Message() string {
	return "server streaming RPCs are not allowed"
}

// Validate implements lint.Rule.
func (r *RPCNoServerStreamingLinked) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range info.File.Services {
		for _, rpc := range service.RPCs {
			if rpc.Response.Stream {
				res = core.AppendIssue(res, r, rpc.Pos, rpc.Name)
			}
		}
	}

	return res, nil
}
