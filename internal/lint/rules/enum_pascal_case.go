package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*EnumPascalCase)(nil)

// EnumPascalCase this rule checks that enums are PascalCase.
type EnumPascalCase struct{}

// Message implements lint.Rule.
func (c *EnumPascalCase) Message() string {
	return "enum name must be in PascalCase"
}

// Validate implements lint.Rule.
func (c *EnumPascalCase) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+(?:[A-Z][a-z]+)*$")
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if !pascalCase.MatchString(enum.EnumName) {
			res = append(res, lint.BuildError(c, enum.Meta.Pos, enum.EnumName))
		}
	}

	return res, nil
}
