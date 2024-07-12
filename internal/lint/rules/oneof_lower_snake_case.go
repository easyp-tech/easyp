package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*OneofLowerSnakeCase)(nil)

// OneofLowerSnakeCase this rule checks that oneof names are lower_snake_case.
type OneofLowerSnakeCase struct{}

// Message implements lint.Rule.
func (c *OneofLowerSnakeCase) Message() string {
	return "oneof name should be lower_snake_case"
}

// Validate implements lint.Rule.
func (c *OneofLowerSnakeCase) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	lowerSnakeCase := regexp.MustCompile("^[a-z]+(_[a-z]+)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, oneof := range message.MessageBody.Oneofs {
			if !lowerSnakeCase.MatchString(oneof.OneofName) {
				res = append(res, lint.BuildError(c, oneof.Meta.Pos, oneof.OneofName))
			}
		}
	}

	return res, nil
}
