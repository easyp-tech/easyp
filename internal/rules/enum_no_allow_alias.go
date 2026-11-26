package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*EnumNoAllowAlias)(nil)

// EnumNoAllowAlias this rule checks that enums do not set allow_alias to true.
type EnumNoAllowAlias struct{}

// Message implements lint.Rule.
func (e *EnumNoAllowAlias) Message() string {
	return "enum must not allow alias"
}

// Validate implements lint.Rule.
func (e *EnumNoAllowAlias) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{Enum: func(enum *protoast.Enum) {
		for _, opt := range enum.Options {
			if opt.Name == "allow_alias" && opt.Value == "true" {
				res = core.AppendIssue(res, e, opt.Pos, enum.Name)
			}
		}
	}})

	return res, nil
}
