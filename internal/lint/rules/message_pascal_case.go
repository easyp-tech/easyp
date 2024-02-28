package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*MessagePascalCase)(nil)

// MessagePascalCase this rule checks that messages are PascalCase.
type MessagePascalCase struct{}

// Validate implements lint.Rule.
func (c *MessagePascalCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+(?:[A-Z][a-z]+)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		if !pascalCase.MatchString(message.MessageName) {
			res = append(res, BuildError(message.Meta.Pos, message.MessageName, lint.ErrMessagePascalCase))
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
