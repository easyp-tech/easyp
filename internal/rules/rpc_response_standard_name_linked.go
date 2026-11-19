package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.LintRule = (*RPCResponseStandardNameLinked)(nil)

// RPCResponseStandardNameLinked is the linked-engine port of
// RPCResponseStandardName (research inventory #30, main-diverged; threat
// T-04-14): the WRITTEN response type string must equal RPCName+"Response"
// or ServiceName+RPCName+"Response" EXACTLY — a qualified written type
// fails both alternatives, and there is no pascal normalization and no
// dot-segment splitting (the lab's last-dot-segment compare would accept
// the qualified form — not ported). The position is the rpc keyword (main
// Meta.Pos ≡ Decl.Pos), never the lab's type-token position.
type RPCResponseStandardNameLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (r *RPCResponseStandardNameLinked) LintRuleName() string {
	return "RPC_RESPONSE_STANDARD_NAME"
}

// Message implements lint.LintRule.
func (r *RPCResponseStandardNameLinked) Message() string {
	return "rpc response should have suffix 'Response'"
}

// Validate implements lint.LintRule.
func (r *RPCResponseStandardNameLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range info.File.Services {
		for _, rpc := range service.RPCs {
			written := rpc.Response.Type
			if written != rpc.Name+"Response" && written != service.Name+rpc.Name+"Response" {
				res = core.AppendLinkedIssue(res, r, rpc.Pos, written)
			}
		}
	}

	return res, nil
}
