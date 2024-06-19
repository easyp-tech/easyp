package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*FieldLowerCamelCase)(nil)

// FieldLowerCamelCase this rule checks that field names are lower_snake_case.
type FieldLowerCamelCase struct{}

// Validate implements lint.Rule.
func (c *FieldLowerCamelCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	lowerSnakeCase := regexp.MustCompile("^[a-z]+([A-Z][a-z]*)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range message.MessageBody.Fields {
			if !lowerSnakeCase.MatchString(field.FieldName) {
				res = append(res, BuildError(field.Meta.Pos, field.FieldName, lint.ErrMessageFieldLowerCamelCase))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
