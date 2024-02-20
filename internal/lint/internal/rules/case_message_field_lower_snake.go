package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*MessageFieldLowerSnakeCase)(nil)

// MessageFieldLowerSnakeCase is a rule for checking fields of messages for lower snake case.
type MessageFieldLowerSnakeCase struct{}

// Validate implements core.Rule.
func (c *MessageFieldLowerSnakeCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	lowerSnakeCase := regexp.MustCompile("^[a-z]+(_[a-z]+)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range message.MessageBody.Fields {
			if !lowerSnakeCase.MatchString(field.FieldName) {
				res = append(res, buildError(field.Meta.Pos, field.FieldName, lint.ErrMessageFieldLowerSnakeCase))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
