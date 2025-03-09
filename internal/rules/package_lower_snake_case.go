package rules

import (
	"regexp"

	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*PackageLowerSnakeCase)(nil)

// PackageLowerSnakeCase his rule checks that packages are lower_snake_case.
type PackageLowerSnakeCase struct{}

// Message implements lint.Rule.
func (c *PackageLowerSnakeCase) Message() string {
	return "package name should be lower_snake_case"
}

// Validate implements lint.Rule.
func (c *PackageLowerSnakeCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	lowerSnakeCase := regexp.MustCompile("^[a-z]+([_|[.][a-z0-9]+)*$")
	for _, pack := range protoInfo.Info.ProtoBody.Packages {
		if !lowerSnakeCase.MatchString(pack.Name) {
			res = core.AppendIssue(res, c, pack.Meta.Pos, pack.Name, pack.Comments)
		}
	}

	return res, nil
}
