package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentOneof)(nil)

// CommentOneof (research
// inventory #37, main-diverged). Presence predicate is MAIN'S: a declaration
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
// Walk shape is main's: oneofs of TOP-LEVEL messages only (direct field
// iteration, no recursion); findings report the declaration keyword position
// (main Meta.Pos ≡ Decl.Pos) with the oneof name as source.
type CommentOneof struct{}

// Message implements lint.Rule.
func (c *CommentOneof) Message() string {
	return "oneof comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentOneof) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, msg := range info.File.Messages {
		for _, oneof := range msg.Oneofs {
			if !oneof.HasCommentAbove() {
				res = core.AppendIssue(res, c, oneof.Pos, oneof.Name)
			}
		}
	}

	return res, nil
}
