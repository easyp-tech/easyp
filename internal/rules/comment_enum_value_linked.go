package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.LintRule = (*CommentEnumValueLinked)(nil)

// CommentEnumValueLinked is the linked-engine port of CommentEnumValue
// (research inventory #34, main-diverged). Presence predicate is MAIN'S: a
// declaration is undocumented iff no comment at all precedes it —
// HasCommentAbove() (Leading+Detached) reproduces main's "no attached
// comment" reading, because go-protoparser's comment parsing attaches ALL
// preceding comments (including across blank lines) to element.Comments and
// keeps the same-line trailing comment in the separate InlineComment field.
// A comment consisting only of a lint directive therefore still documents
// the declaration, exactly as in main; the engine's central machinery owns
// suppression (rules never read comment content, never filter directives).
// The lab's directive-filtering helper is deliberately not ported — its
// directive-only => undocumented reading would fabricate findings on
// suppression-heavy corpora.
//
// Walk shape is main's: values of top-level enums plus the values of enums
// of top-level messages (direct field iteration, no recursion); findings
// report the value's first-token position (main Meta.Pos ≡ Decl.Pos) with
// the value name as source.
type CommentEnumValueLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (c *CommentEnumValueLinked) LintRuleName() string {
	return "COMMENT_ENUM_VALUE"
}

// Message implements lint.LintRule.
func (c *CommentEnumValueLinked) Message() string {
	return "enum value comments must not be empty"
}

// Validate implements lint.LintRule.
func (c *CommentEnumValueLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, enum := range info.File.Enums {
		for _, value := range enum.Values {
			if !value.HasCommentAbove() {
				res = core.AppendLinkedIssue(res, c, value.Pos, value.Name)
			}
		}
	}

	for _, msg := range info.File.Messages {
		for _, enum := range msg.Enums {
			for _, value := range enum.Values {
				if !value.HasCommentAbove() {
					res = core.AppendLinkedIssue(res, c, value.Pos, value.Name)
				}
			}
		}
	}

	return res, nil
}
