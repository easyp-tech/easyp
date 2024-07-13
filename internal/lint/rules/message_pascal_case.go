package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*MessagePascalCase)(nil)

// MessagePascalCase this rule checks that messages are PascalCase.
type MessagePascalCase struct{}

// Message implements lint.Rule.
func (c *MessagePascalCase) Message() string {
	return "message name should be PascalCase"
}

// Validate implements lint.Rule.
func (c *MessagePascalCase) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+(?:[A-Z][a-z]+)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		if !pascalCase.MatchString(message.MessageName) {
			res = lint.AppendIssue(res, c, message.Meta.Pos, message.MessageName, message.Comments)
		}
	}

	return res, nil
}
