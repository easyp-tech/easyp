package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*EnumPascalCase)(nil)

// EnumPascalCase (research
// inventory #7, main-diverged): main's regexp wins over the lab's
// casing-helper equality (casing.go) — acronym enums like "ABC" fail main but
// would pass the helper — and the position is the declaration keyword
// (enum.Meta.Pos ≡ Decl.Pos), not the lab's declared-name token. Walk shape:
// main's two-level loop — top-level enums plus enums of TOP-LEVEL messages;
// deeper nesting is outside main's predicate, so the recursion-capable
// Visitor walk is not used.
type EnumPascalCase struct{}

// Message implements lint.Rule.
func (c *EnumPascalCase) Message() string {
	return "enum name must be in PascalCase"
}

// Validate implements lint.Rule.
func (c *EnumPascalCase) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	pascalCase := regexp.MustCompile("^[A-Z][a-z]+(?:[A-Z][a-z]+)*$")

	for _, enum := range info.File.Enums {
		if !pascalCase.MatchString(enum.Name) {
			// enum.Pos is the declaration's first token — the same position
			// main reported as Meta.Pos (pinned by TestDeclPositionProbe).
			res = core.AppendIssue(res, c, enum.Pos, enum.Name)
		}
	}

	for _, msg := range info.File.Messages {
		for _, enum := range msg.Enums {
			if !pascalCase.MatchString(enum.Name) {
				res = core.AppendIssue(res, c, enum.Pos, enum.Name)
			}
		}
	}

	return res, nil
}
