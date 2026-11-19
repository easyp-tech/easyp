package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.LintRule = (*MessagePascalCaseLinked)(nil)

// MessagePascalCaseLinked is the linked-engine port of MessagePascalCase
// (research inventory #13, main-diverged): main's regexp wins over the lab's
// casing-helper equality (casing.go) — a single-char message ("A") fails main
// but would pass the helper — and the position is the declaration keyword
// (message.Meta.Pos ≡ Decl.Pos, main's test pins column 1), not the lab's
// declared-name token.
// Walk shape: TOP-LEVEL messages only (main's shallow loop); nested messages
// are outside main's predicate, so the recursion-capable Visitor walk is not
// used.
type MessagePascalCaseLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (c *MessagePascalCaseLinked) LintRuleName() string {
	return "MESSAGE_PASCAL_CASE"
}

// Message implements lint.LintRule.
func (c *MessagePascalCaseLinked) Message() string {
	return "message name should be PascalCase"
}

// Validate implements lint.LintRule.
func (c *MessagePascalCaseLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
	var res []core.Issue

	pascalCase := regexp.MustCompile("^[A-Z][a-zA-Z0-9]+(?:[A-Z][a-zA-Z0-9]+)*$")

	for _, message := range info.File.Messages {
		if !pascalCase.MatchString(message.Name) {
			// message.Pos is the declaration's first token — the same position
			// main reported as Meta.Pos (pinned by TestDeclPositionProbe).
			res = core.AppendLinkedIssue(res, c, message.Pos, message.Name)
		}
	}

	return res, nil
}
