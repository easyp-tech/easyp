package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentServiceLinked)(nil)

// CommentServiceLinked is the linked-engine port of CommentService (research
// inventory #39, main-diverged). Presence predicate is MAIN'S: a declaration
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
// Walk shape is main's: top-level services only (direct field iteration);
// findings report the declaration keyword position (main Meta.Pos ≡
// Decl.Pos) with the service name as source.
type CommentServiceLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (c *CommentServiceLinked) LintRuleName() string {
	return "COMMENT_SERVICE"
}

// Message implements lint.Rule.
func (c *CommentServiceLinked) Message() string {
	return "service comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentServiceLinked) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range info.File.Services {
		if !service.HasCommentAbove() {
			res = core.AppendIssue(res, c, service.Pos, service.Name)
		}
	}

	return res, nil
}
