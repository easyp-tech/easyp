package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageLowerSnakeCase)(nil)

// PackageLowerSnakeCase his rule checks that packages are lower_snake_case.
type PackageLowerSnakeCase struct{}

// Message implements lint.Rule.
func (c *PackageLowerSnakeCase) Message() string {
	return "package name should be lower_snake_case"
}

// Validate implements lint.Rule.
func (c *PackageLowerSnakeCase) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	lowerSnakeCase := regexp.MustCompile("^[a-z]+([_|[.][a-z0-9]+)*$")
	for _, pack := range protoInfo.Info.ProtoBody.Packages {
		if !lowerSnakeCase.MatchString(pack.Name) {
			res = append(res, lint.BuildError(c, pack.Meta.Pos, pack.Name))
		}
	}

	return res, nil
}
