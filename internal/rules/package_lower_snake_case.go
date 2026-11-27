package rules

import (
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageLowerSnakeCase)(nil)

// PackageLowerSnakeCase this rule checks that every package component is lower_snake_case.
type PackageLowerSnakeCase struct{}

// Message implements lint.Rule.
func (c *PackageLowerSnakeCase) Message() string {
	return "package name should be lower_snake_case"
}

// Validate implements lint.Rule.
func (c *PackageLowerSnakeCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	for _, pkg := range protoInfo.File.Packages {
		parts := strings.Split(pkg.Name, ".")
		for i, part := range parts {
			parts[i] = toLowerSnakeCase(part)
		}
		if pkg.Name != strings.Join(parts, ".") {
			res = core.AppendIssue(res, c, pkg.Pos, pkg.Name)
		}
	}

	return res, nil
}
