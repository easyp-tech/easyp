package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
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
	upperSnakeCase := regexp.MustCompile("^[A-Z]+(_[A-Z]+)*$")
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		for _, field := range enum.EnumBody.EnumFields {
			if !upperSnakeCase.MatchString(field.Ident) {
				res = core.AppendIssue(res, c, field.Meta.Pos, field.Ident, field.Comments)
			}
		}
	}

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, enum := range msg.MessageBody.Enums {
			for _, field := range enum.EnumBody.EnumFields {
				if !upperSnakeCase.MatchString(field.Ident) {
					res = core.AppendIssue(res, c, field.Meta.Pos, field.Ident, field.Comments)
				}
			}
		}
	}

	return res, nil
}
