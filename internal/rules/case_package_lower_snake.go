package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageLowerSnakeCase)(nil)

// PackageLowerSnakeCase is a rule for checking package for lower snake case.
type PackageLowerSnakeCase struct{}

// Validate implements Rule.
func (c *PackageLowerSnakeCase) Validate(protoInfo core.ProtoInfo) []error {
	var res []error
	lowerSnakeCase := regexp.MustCompile("^[a-z]+([_|[.][a-z0-9]+)*$")
	for _, pack := range protoInfo.Info.ProtoBody.Packages {
		if !lowerSnakeCase.MatchString(pack.Name) {
			res = append(res, buildError(pack.Meta.Pos, pack.Name, core.ErrPackageLowerSnakeCase))
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
