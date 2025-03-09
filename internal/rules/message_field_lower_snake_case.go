package rules

import (
	"regexp"

	"go.redsock.ru/protopack/internal/core"
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

	lowerSnakeCase := regexp.MustCompile("^[a-z0-9]+(_[a-z0-9]+)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range message.MessageBody.Fields {
			if !lowerSnakeCase.MatchString(field.FieldName) {
				res = core.AppendIssue(res, c, field.Meta.Pos, field.FieldName, field.Comments)
			}
		}
	}

	return res, nil
}
