package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*EnumValueUpperSnakeCase)(nil)

// EnumValueUpperSnakeCase this rule checks that enum values are UPPER_SNAKE_CASE.
type EnumValueUpperSnakeCase struct{}

// Message implements lint.Rule.
func (c *EnumValueUpperSnakeCase) Message() string {
	return "enum value must be in UPPER_SNAKE_CASE"
}

// Validate implements lint.Rule.
func (c *EnumValueUpperSnakeCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{EnumValue: func(_ *protoast.Enum, value *protoast.EnumValue) {
		if value.Name != toUpperSnakeCase(value.Name) {
			res = core.AppendIssue(res, c, value.NamePos, value.Name)
		}
	}})

	return res, nil
}
