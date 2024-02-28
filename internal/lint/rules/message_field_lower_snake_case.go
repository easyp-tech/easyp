package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*FieldLowerSnakeCase)(nil)

// FieldLowerSnakeCase this rule checks that field names are lower_snake_case.
type FieldLowerSnakeCase struct{}

// Validate implements lint.Rule.
func (c *FieldLowerSnakeCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	lowerSnakeCase := regexp.MustCompile("^[a-z]+(_[a-z]+)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range message.MessageBody.Fields {
			if !lowerSnakeCase.MatchString(field.FieldName) {
				res = append(res, BuildError(field.Meta.Pos, field.FieldName, lint.ErrMessageFieldLowerSnakeCase))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
