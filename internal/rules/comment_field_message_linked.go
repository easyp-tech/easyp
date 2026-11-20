package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentField)(nil)

// CommentField (research
// inventory #35, main-diverged). Presence predicate is MAIN'S: a declaration
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
// Walk shape is main's: plain fields of TOP-LEVEL messages only (direct
// field iteration, no recursion, no oneof members). Proto2 group fields need
// no guard: both models list groups separately from the plain-field list
// (go-protoparser MessageBody.Groups, protoast Message.Groups), so they never
// enter this loop — research A4 resolved at source. Findings report the
// field's first-token position (main Meta.Pos ≡ Decl.Pos, the type token)
// with the field name as source.
type CommentField struct{}

// Message implements lint.Rule.
func (c *CommentField) Message() string {
	return "field comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentField) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, message := range info.File.Messages {
		for _, field := range message.Fields {
			if !field.HasCommentAbove() {
				res = core.AppendIssue(res, c, field.Pos, field.Name)
			}
		}
	}

	return res, nil
}
