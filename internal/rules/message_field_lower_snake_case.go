package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*FieldLowerSnakeCase)(nil)

// FieldLowerSnakeCase this rule checks that field names are lower_snake_case.
type FieldLowerSnakeCase struct{}

// Message implements lint.Rule.
func (c *FieldLowerSnakeCase) Message() string {
	return "message field should be lower_snake_case"
}

// Validate implements lint.Rule.
func (c *FieldLowerSnakeCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{Field: func(field *protoast.Field) {
		if field.Name != toLowerSnakeCase(field.Name) {
			res = core.AppendIssue(res, c, field.NamePos, field.Name)
		}
	}})

	return res, nil
}
