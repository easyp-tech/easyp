package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

// EnumNoAllowAlias this rule checks that enums are PascalCase.
type EnumNoAllowAlias struct{}

// Message implements lint.Rule.
func (e *EnumNoAllowAlias) Message() string {
	return "enum must not allow alias"
}

// Validate implements lint.Rule.
func (e *EnumNoAllowAlias) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		for _, opt := range enum.EnumBody.Options {
			if opt.OptionName == "allow_alias" {
				res = lint.AppendIssue(res, e, enum.Meta.Pos, enum.EnumName, enum.Comments)
			}
		}
	}

	return res, nil
}
