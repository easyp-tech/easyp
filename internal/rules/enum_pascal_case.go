package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
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
	protoInfo.File.Walk(protoast.Visitor{Enum: func(enum *protoast.Enum) {
		if enum.Name != toPascalCase(enum.Name) {
			res = core.AppendIssue(res, c, enum.NamePos, enum.Name)
		}
	}})

	return res, nil
}
