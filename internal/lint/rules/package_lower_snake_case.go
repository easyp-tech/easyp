package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageLowerSnakeCase)(nil)

// PackageLowerSnakeCase his rule checks that packages are lower_snake_case.
type PackageLowerSnakeCase struct{}

// Validate implements lint.Rule.
func (c *PackageLowerSnakeCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	lowerSnakeCase := regexp.MustCompile("^[a-z]+([_|[.][a-z0-9]+)*$")
	for _, pack := range protoInfo.Info.ProtoBody.Packages {
		if !lowerSnakeCase.MatchString(pack.Name) {
			res = AppendError(res, PACKAGE_LOWER_SNAKE_CASE, pack.Meta.Pos, pack.Name, pack.Comments)
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
