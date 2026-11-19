package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.LintRule = (*PackageLowerSnakeCaseLinked)(nil)

// PackageLowerSnakeCaseLinked is the linked-engine port of
// PackageLowerSnakeCase (research inventory #15, main-diverged): main's
// single regexp wins over the lab's per-component casing-helper equality —
// digits fail main in the first component ("foo1bar") but pass the helper —
// and the position is the declaration keyword (pack.Meta.Pos ≡ Decl.Pos).
type PackageLowerSnakeCaseLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (c *PackageLowerSnakeCaseLinked) LintRuleName() string {
	return "PACKAGE_LOWER_SNAKE_CASE"
}

// Message implements lint.LintRule.
func (c *PackageLowerSnakeCaseLinked) Message() string {
	return "package name should be lower_snake_case"
}

// Validate implements lint.LintRule.
func (c *PackageLowerSnakeCaseLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
	var res []core.Issue

	lowerSnakeCase := regexp.MustCompile("^[a-z]+([_|[.][a-z0-9]+)*$")

	for _, pack := range info.File.Packages {
		if !lowerSnakeCase.MatchString(pack.Name) {
			res = core.AppendLinkedIssue(res, c, pack.Pos, pack.Name)
		}
	}

	return res, nil
}
