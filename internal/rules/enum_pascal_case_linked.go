package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.LintRule = (*EnumPascalCaseLinked)(nil)

// EnumPascalCaseLinked is the linked-engine port of EnumPascalCase (research
// inventory #7, main-diverged): main's regexp wins over the lab's
// casing-helper equality (casing.go) — acronym enums like "ABC" fail main but
// would pass the helper — and the position is the declaration keyword
// (enum.Meta.Pos ≡ Decl.Pos), not the lab's declared-name token. Walk shape:
// main's two-level loop — top-level enums plus enums of TOP-LEVEL messages;
// deeper nesting is outside main's predicate, so the recursion-capable
// Visitor walk is not used.
type EnumPascalCaseLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (c *EnumPascalCaseLinked) LintRuleName() string {
	return "ENUM_PASCAL_CASE"
}

// Message implements lint.LintRule.
func (c *EnumPascalCaseLinked) Message() string {
	return "enum name must be in PascalCase"
}

// Validate implements lint.LintRule.
func (c *EnumPascalCaseLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
	var res []core.Issue

	pascalCase := regexp.MustCompile("^[A-Z][a-z]+(?:[A-Z][a-z]+)*$")

	for _, enum := range info.File.Enums {
		if !pascalCase.MatchString(enum.Name) {
			// enum.Pos is the declaration's first token — the same position
			// main reported as Meta.Pos (pinned by TestDeclPositionProbe).
			res = core.AppendLinkedIssue(res, c, enum.Pos, enum.Name)
		}
	}

	for _, msg := range info.File.Messages {
		for _, enum := range msg.Enums {
			if !pascalCase.MatchString(enum.Name) {
				res = core.AppendLinkedIssue(res, c, enum.Pos, enum.Name)
			}
		}
	}

	return res, nil
}
