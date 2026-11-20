package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.ModuleRule = (*RPCRequestResponseUniqueLinked)(nil)

// RPCRequestResponseUniqueLinked is the linked-engine port of
// RPCRequestResponseUnique (research inventory #28, main-diverged + adapt;
// threat T-04-13): main is PER-FILE uniqueness on the WRITTEN type strings
// with the first user exempt — later users issue at the rpc keyword
// position (main Meta.Pos ≡ Decl.Pos), and input==output of one rpc counts
// as a duplicate within the file. The etalon contributes the Collect/Finish
// ModuleRule lifecycle; its cross-file qualified-name and report-everyone
// semantics are NOT ported: Finish groups the accumulated occurrences by
// file and reproduces main's within-file first-exempt (lo.Contains) logic
// on the written names. The engine-provided resolution table is never read
// (the linked-file resolution pointer stays unused for parity — grep-clean).
//
// Collect runs in the engine's serial drain in walk order; Finish reports
// once after the last file and resets the receiver state.
type RPCRequestResponseUniqueLinked struct {
	occurrences []uniqueOccurrence
}

// uniqueOccurrence is one written type use: the request or response of one
// rpc, in walk order.
type uniqueOccurrence struct {
	path string
	pos  core.Position
	name string
}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (r *RPCRequestResponseUniqueLinked) LintRuleName() string {
	return "RPC_REQUEST_RESPONSE_UNIQUE"
}

// Message implements lint.Rule.
func (r *RPCRequestResponseUniqueLinked) Message() string {
	return "request and response types must be unique across all RPCs"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (r *RPCRequestResponseUniqueLinked) Validate(core.ProtoInfo) ([]core.Issue, error) {
	return nil, nil
}

// Collect implements core.ModuleRule.
func (r *RPCRequestResponseUniqueLinked) Collect(info core.ProtoInfo) error {
	for _, service := range info.File.Services {
		for _, rpc := range service.RPCs {
			// rpc.Pos is the rpc keyword — the same position main reported as
			// Meta.Pos. The written type strings are compared as written; the
			// resolution table is not consulted.
			r.occurrences = append(r.occurrences,
				uniqueOccurrence{
					path: info.Path,
					pos:  rpc.Pos,
					name: rpc.Request.Type,
				},
				uniqueOccurrence{
					path: info.Path,
					pos:  rpc.Pos,
					name: rpc.Response.Type,
				},
			)
		}
	}

	return nil
}

// Finish implements core.ModuleRule.
func (r *RPCRequestResponseUniqueLinked) Finish() []core.IssueInfo {
	var res []core.IssueInfo

	var order []string
	byPath := make(map[string][]uniqueOccurrence)
	for _, occ := range r.occurrences {
		if _, ok := byPath[occ.path]; !ok {
			order = append(order, occ.path)
		}
		byPath[occ.path] = append(byPath[occ.path], occ)
	}

	for _, path := range order {
		seen := make(map[string]struct{})
		for _, occ := range byPath[path] {
			if _, ok := seen[occ.name]; !ok {
				// first user of the written name within this file: exempt.
				seen[occ.name] = struct{}{}
				continue
			}
			for _, issue := range core.AppendIssue(nil, r, occ.pos, occ.name) {
				res = append(res, core.IssueInfo{Issue: issue, Path: occ.path})
			}
		}
	}

	r.occurrences = nil

	return res
}
