package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*EnumValueUpperSnakeCase)(nil)

// EnumValueUpperSnakeCase is a rule for checking value of enum for upper snake case.
type EnumValueUpperSnakeCase struct{}

// Validate implements core.Rule.
func (c *EnumValueUpperSnakeCase) Validate(protoInfo core.ProtoInfo) []error {
	var res []error
	upperSnakeCase := regexp.MustCompile("^[A-Z]+(_[A-Z]+)*$")
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		for _, field := range enum.EnumBody.EnumFields {
			if !upperSnakeCase.MatchString(field.Ident) {
				res = append(res, buildError(field.Meta.Pos, field.Ident, core.ErrEnumValueUpperSnakeCase))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
