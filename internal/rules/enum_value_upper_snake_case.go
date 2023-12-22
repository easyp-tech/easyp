package rules

import (
	"fmt"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"regexp"
)

var _ core.Rule = (*EnumValueUpperSnakeCase)(nil)

// EnumValueUpperSnakeCase is a rule for checking value of enum for upper snake case.
type EnumValueUpperSnakeCase struct{}

// Validate implements Rule.
func (c *EnumValueUpperSnakeCase) Validate(svc *unordered.Proto) []error {
	var res []error
	upperSnakeCase := regexp.MustCompile("^[A-Z]+(_[A-Z]+)*$")
	for _, enum := range svc.ProtoBody.Enums {
		for _, field := range enum.EnumBody.EnumFields {
			if !upperSnakeCase.MatchString(field.Ident) {
				res = append(res, &Error{
					Err: fmt.Errorf("%d:%d:%s: %w", field.Meta.Pos.Line, field.Meta.Pos.Column, field.Ident, core.ErrEnumValueUpperSnakeCase),
				})
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
