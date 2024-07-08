package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*OneofLowerCamelCase)(nil)

// OneofLowerCamelCase this rule checks that oneof names are lowerCamelcase.
type OneofLowerCamelCase struct{}

// Validate implements lint.Rule.
func (c *OneofLowerCamelCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	lowerCamelCase := regexp.MustCompile("^[a-z]+([A-Z][a-z]*)*$")
	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, oneof := range message.MessageBody.Oneofs {
			if !lowerCamelCase.MatchString(oneof.OneofName) {
				res = append(res, BuildError(oneof.Meta.Pos, oneof.OneofName, lint.ErrOneofLowerCamelCase))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
