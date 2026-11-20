package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentRPC)(nil)

// CommentRPC (research
// inventory #38, main-diverged). Presence predicate is MAIN'S: a declaration
// is undocumented iff no comment at all precedes it — HasCommentAbove()
// (Leading+Detached) reproduces main's "no attached comment" reading,
// because go-protoparser's comment parsing attaches ALL preceding comments
// (including across blank lines) to element.Comments and keeps the same-line
// trailing comment in the separate InlineComment field. A comment consisting
// only of a lint directive therefore still documents the declaration,
// exactly as in main; the engine's central machinery owns suppression (rules
// never read comment content, never filter directives). The lab's
// directive-filtering helper is deliberately not ported — its
// directive-only => undocumented reading would fabricate findings on
// suppression-heavy corpora.
//
// Walk shape is main's: RPCs of top-level services only (direct field
// iteration); findings report the rpc keyword position (main Meta.Pos ≡
// Decl.Pos) with the rpc name as source.
type CommentRPC struct{}

// Message implements lint.Rule.
func (c *CommentRPC) Message() string {
	return "rpc comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentRPC) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range info.File.Services {
		for _, rpc := range service.RPCs {
			if !rpc.HasCommentAbove() {
				res = core.AppendIssue(res, c, rpc.Pos, rpc.Name)
			}
		}
	}

	return res, nil
}
