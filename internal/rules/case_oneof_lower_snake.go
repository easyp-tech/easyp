package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*OneofLowerSnakeCase)(nil)

// OneofLowerSnakeCase is a rule for checking oneof of messages for lower snake case.
type OneofLowerSnakeCase struct{}

// Validate implements core.Rule.
func (c *OneofLowerSnakeCase) Validate(protoInfo core.ProtoInfo) []error {
	var res []error
	lowerSnakeCase := regexp.MustCompile("^[a-z]+(_[a-z]+)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, oneof := range message.MessageBody.Oneofs {
			if !lowerSnakeCase.MatchString(oneof.OneofName) {
				res = append(res, buildError(oneof.Meta.Pos, oneof.OneofName, core.ErrOneofLowerSnakeCase))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
