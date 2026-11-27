package rules

import (
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*EnumValuePrefix)(nil)

// EnumValuePrefix this rule checks that enum values are prefixed with ENUM_NAME_UPPER_SNAKE_CASE.
type EnumValuePrefix struct{}

// Message implements lint.Rule.
func (e *EnumValuePrefix) Message() string {
	return "enum value prefix is not valid"
}

// Validate implements lint.Rule.
func (e *EnumValuePrefix) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{EnumValue: func(enum *protoast.Enum, value *protoast.EnumValue) {
		if !strings.HasPrefix(value.Name, toUpperSnakeCase(enum.Name)+"_") {
			res = core.AppendIssue(res, e, value.NamePos, value.Name)
		}
	}})

	return res, nil
}
