package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*EnumValueUpperSnakeCase)(nil)

// EnumValueUpperSnakeCase this rule checks that enum values are UPPER_SNAKE_CASE.
type EnumValueUpperSnakeCase struct{}

// Message implements lint.Rule.
func (c *EnumValueUpperSnakeCase) Message() string {
	return "enum value must be in UPPER_SNAKE_CASE"
}

// Validate implements lint.Rule.
func (c *EnumValueUpperSnakeCase) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	upperSnakeCase := regexp.MustCompile("^[A-Z]+(_[A-Z]+)*$")
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		for _, field := range enum.EnumBody.EnumFields {
			if !upperSnakeCase.MatchString(field.Ident) {
				res = append(res, lint.BuildError(c, field.Meta.Pos, field.Ident))
			}
		}
	}

	return res, nil
}
