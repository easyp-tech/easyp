package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*FieldLowerSnakeCase)(nil)

// FieldLowerSnakeCase this rule checks that field names are lower_snake_case.
type FieldLowerSnakeCase struct{}

// Message implements lint.Rule.
func (c *FieldLowerSnakeCase) Message() string {
	return "message field should be lower_snake_case"
}

// Validate implements lint.Rule.
func (c *FieldLowerSnakeCase) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	lowerSnakeCase := regexp.MustCompile("^[a-z]+(_[a-z]+)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range message.MessageBody.Fields {
			if !lowerSnakeCase.MatchString(field.FieldName) {
				res = append(res, lint.BuildError(field.Meta.Pos, field.FieldName, c.Message()))
			}
		}
	}

	return res, nil
}
