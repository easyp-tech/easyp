package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*MessagePascalCase)(nil)

// MessagePascalCase this rule checks that messages are PascalCase.
type MessagePascalCase struct{}

// Message implements lint.Rule.
func (c *MessagePascalCase) Message() string {
	return "message name should be PascalCase"
}

// Validate implements lint.Rule.
func (c *MessagePascalCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	pascalCase := regexp.MustCompile("^[A-Z][a-zA-Z0-9]+(?:[A-Z][a-zA-Z0-9]+)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		if !pascalCase.MatchString(message.MessageName) {
			res = core.AppendIssue(res, c, message.Meta.Pos, message.MessageName, message.Comments)
		}
	}

	return res, nil
}
