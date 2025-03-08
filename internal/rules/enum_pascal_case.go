package rules

import (
	"regexp"

	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*EnumPascalCase)(nil)

// EnumPascalCase this rule checks that enums are PascalCase.
type EnumPascalCase struct{}

// Message implements lint.Rule.
func (c *EnumPascalCase) Message() string {
	return "enum name must be in PascalCase"
}

// Validate implements lint.Rule.
func (c *EnumPascalCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+(?:[A-Z][a-z]+)*$")
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if !pascalCase.MatchString(enum.EnumName) {
			res = core.AppendIssue(res, c, enum.Meta.Pos, enum.EnumName, enum.Comments)
		}
	}

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, enum := range msg.MessageBody.Enums {
			if !pascalCase.MatchString(enum.EnumName) {
				res = core.AppendIssue(res, c, enum.Meta.Pos, enum.EnumName, enum.Comments)
			}
		}
	}

	return res, nil
}
