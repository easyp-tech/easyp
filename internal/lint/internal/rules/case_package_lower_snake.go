package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageLowerSnakeCase)(nil)

// PackageLowerSnakeCase is a rule for checking package for lower snake case.
type PackageLowerSnakeCase struct{}

// Validate implements core.Rule.
func (c *PackageLowerSnakeCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	lowerSnakeCase := regexp.MustCompile("^[a-z]+([_|[.][a-z0-9]+)*$")
	for _, pack := range protoInfo.Info.ProtoBody.Packages {
		if !lowerSnakeCase.MatchString(pack.Name) {
			res = append(res, buildError(pack.Meta.Pos, pack.Name, lint.ErrPackageLowerSnakeCase))
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
