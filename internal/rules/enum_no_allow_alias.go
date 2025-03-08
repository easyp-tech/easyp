package rules

import (
	"go.redsock.ru/protopack/internal/core"
)

// EnumNoAllowAlias this rule checks that enums are PascalCase.
type EnumNoAllowAlias struct{}

// Message implements lint.Rule.
func (e *EnumNoAllowAlias) Message() string {
	return "enum must not allow alias"
}

// Validate implements lint.Rule.
func (e *EnumNoAllowAlias) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		for _, opt := range enum.EnumBody.Options {
			if opt.OptionName == "allow_alias" {
				res = core.AppendIssue(res, e, enum.Meta.Pos, enum.EnumName, enum.Comments)
			}
		}
	}

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, enum := range msg.MessageBody.Enums {
			for _, opt := range enum.EnumBody.Options {
				if opt.OptionName == "allow_alias" {
					res = core.AppendIssue(res, e, enum.Meta.Pos, enum.EnumName, enum.Comments)
				}
			}
		}
	}

	return res, nil
}
