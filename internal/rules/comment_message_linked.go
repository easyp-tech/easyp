package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentMessage)(nil)

// CommentMessage (research
// inventory #36, main-diverged). Presence predicate is MAIN'S: a declaration
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
// Walk shape is main's: top-level messages only (direct field iteration, no
// recursion into nested messages); findings report the declaration keyword
// position (main Meta.Pos ≡ Decl.Pos) with the message name as source.
type CommentMessage struct{}

// Message implements lint.Rule.
func (c *CommentMessage) Message() string {
	return "message comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentMessage) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, message := range info.File.Messages {
		if !message.HasCommentAbove() {
			res = core.AppendIssue(res, c, message.Pos, message.Name)
		}
	}

	return res, nil
}
