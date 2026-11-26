package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageLowerSnakeCase)(nil)

// PackageLowerSnakeCase (research inventory #15, main-diverged): main's
// single regexp wins over the lab's per-component casing-helper equality —
// digits fail main in the first component ("foo1bar") but pass the helper —
// and the position is the declaration keyword (pack.Meta.Pos ≡ Decl.Pos).
type PackageLowerSnakeCase struct{}

// Message implements lint.Rule.
func (c *PackageLowerSnakeCase) Message() string {
	return "package name should be lower_snake_case"
}

// Validate implements lint.Rule.
func (c *PackageLowerSnakeCase) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	lowerSnakeCase := regexp.MustCompile("^[a-z]+([_|[.][a-z0-9]+)*$")

	for _, pack := range info.File.Packages {
		if !lowerSnakeCase.MatchString(pack.Name) {
			res = core.AppendIssue(res, c, pack.Pos, pack.Name)
		}
	}

	return res, nil
}
