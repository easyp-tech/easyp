package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*OneofLowerSnakeCase)(nil)

// OneofLowerSnakeCase this rule checks that oneof names are lower_snake_case.
type OneofLowerSnakeCase struct{}

// Validate implements lint.Rule.
func (c *OneofLowerSnakeCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	lowerSnakeCase := regexp.MustCompile("^[a-z]+(_[a-z]+)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, oneof := range message.MessageBody.Oneofs {
			if !lowerSnakeCase.MatchString(oneof.OneofName) {
				res = append(res, BuildError(oneof.Meta.Pos, oneof.OneofName, ErrOneofLowerSnakeCase))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
