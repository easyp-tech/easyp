package rules

import (
	"regexp"

	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*OneofLowerSnakeCase)(nil)

// OneofLowerSnakeCase this rule checks that oneof names are lower_snake_case.
type OneofLowerSnakeCase struct{}

// Message implements lint.Rule.
func (c *OneofLowerSnakeCase) Message() string {
	return "oneof name should be lower_snake_case"
}

// Validate implements lint.Rule.
func (c *OneofLowerSnakeCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	lowerSnakeCase := regexp.MustCompile("^[a-z]+(_[a-z]+)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, oneof := range message.MessageBody.Oneofs {
			if !lowerSnakeCase.MatchString(oneof.OneofName) {
				res = core.AppendIssue(res, c, oneof.Meta.Pos, oneof.OneofName, oneof.Comments)
			}
		}
	}

	return res, nil
}
