package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*EnumPascalCase)(nil)

// EnumPascalCase is a rule for checking name of enum for pascal case.
type EnumPascalCase struct{}

// Validate implements core.Rule.
func (c *EnumPascalCase) Validate(protoInfo core.ProtoInfo) []error {
	var res []error
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+(?:[A-Z][a-z]+)*$")
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if !pascalCase.MatchString(enum.EnumName) {
			res = append(res, buildError(enum.Meta.Pos, enum.EnumName, core.ErrEnumPascalCase))
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
