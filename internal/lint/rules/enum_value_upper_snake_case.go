package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*EnumValueUpperSnakeCase)(nil)

// EnumValueUpperSnakeCase this rule checks that enum values are UPPER_SNAKE_CASE.
type EnumValueUpperSnakeCase struct{}

// Validate implements lint.Rule.
func (c *EnumValueUpperSnakeCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	upperSnakeCase := regexp.MustCompile("^[A-Z]+(_[A-Z]+)*$")
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		for _, field := range enum.EnumBody.EnumFields {
			if !upperSnakeCase.MatchString(field.Ident) {
				res = append(res, BuildError(field.Meta.Pos, field.Ident, lint.ErrEnumValueUpperSnakeCase))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
