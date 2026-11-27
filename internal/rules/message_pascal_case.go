package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
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
	protoInfo.File.Walk(protoast.Visitor{Message: func(message *protoast.Message) {
		if message.Name != toPascalCase(message.Name) {
			res = core.AppendIssue(res, c, message.NamePos, message.Name)
		}
	}})

	return res, nil
}
